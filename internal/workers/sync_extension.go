package workers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/cli"
	"github.com/vscodethemes/backend/internal/colors"
	"github.com/vscodethemes/backend/internal/db"
	"github.com/vscodethemes/backend/internal/downloader"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
	"golang.org/x/sync/errgroup"
)

type SyncExtensionArgs struct {
	ExtensionName string `json:"extensionName"`
	PublisherName string `json:"publisherName"`
	Force         bool   `json:"force"`
}

func (SyncExtensionArgs) Kind() string {
	return "syncExtension"
}

func (SyncExtensionArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       SyncExtensionPriorityQueue,
		MaxAttempts: 5,
	}
}

type SyncExtensionWorker struct {
	river.WorkerDefaults[SyncExtensionArgs]
	Marketplace       *marketplace.Client
	Directory         string
	DisableCleanup    bool
	ObjectStoreClient *s3.Client
	ObjectStoreBucket string
	CDNBaseUrl        string
	DBPool            *pgxpool.Pool
}

func (w *SyncExtensionWorker) Timeout(*river.Job[SyncExtensionArgs]) time.Duration {
	return 5 * time.Minute
}

func (w *SyncExtensionWorker) Work(ctx context.Context, job *river.Job[SyncExtensionArgs]) error {
	extensionSlug := fmt.Sprintf("%s.%s", job.Args.PublisherName, job.Args.ExtensionName)
	log.Infof("Syncing extension package: %s", extensionSlug)

	// Add a delay to avoid rate limiting from the martketplace API.
	time.Sleep(2 * time.Second)

	// Fetch extension from the marketplace API.
	queryResults, err := w.Marketplace.NewQuery(ctx, qo.WithSlug(extensionSlug))
	if err != nil {
		return fmt.Errorf("failed to query marketplace: %w", err)
	}

	if len(queryResults) == 0 {
		return fmt.Errorf("extension not found")
	}

	extension := queryResults[0]

	isUpToDate, err := isExtensionUpToDate(ctx, db.New(w.DBPool), extension)
	if err != nil {
		return fmt.Errorf("failed to check if extension is up to date: %w", err)
	}

	if isUpToDate && !job.Args.Force {
		log.Infof("Extension is up to date, skipping")
		return nil
	}

	upsertExtensionParams, err := convertUpsertExtensionParams(extension)
	if err != nil {
		return fmt.Errorf("failed to convert upsert extension params: %w", err)
	}

	// Ensure there's a package URL for the extension.
	packageUrl := extension.GetPackageURL()
	if packageUrl == "" {
		return fmt.Errorf("extension package not found")
	}

	// Create a directory for the job to download the package.
	jobDir := path.Join(w.Directory, "jobs", fmt.Sprintf("%d", job.ID))
	err = os.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create job dir: %w", err)
	}
	if !w.DisableCleanup {
		defer func() {
			log.Infof("Cleaning up job directory: %s", jobDir)
			os.RemoveAll(jobDir)
		}()
	}

	// Download the extension package.
	d := downloader.New(jobDir, extensionSlug)

	log.Infof("Downloading package: %s", packageUrl)
	err = d.Download(ctx, packageUrl)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	log.Infof("Extracting package: %s", d.PackagePath)
	err = d.Extract()
	if err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	extensionPath, err := filepath.Abs(d.ExtractDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for extension: %w", err)
	}

	log.Infof("Reading extension info: %s", extensionPath)
	info, err := cli.GetInfo(ctx, extensionPath)
	if err != nil {
		return fmt.Errorf("failed to get info: %w", err)
	}

	imagesPath, err := filepath.Abs(path.Join(jobDir, "images"))
	if err != nil {
		return fmt.Errorf("failed to get absolute path for images: %w", err)
	}

	// Generate images for each theme concurrency, up to a max of 10 subroutines.
	imagesResults := []cli.GenerateImagesResult{}
	group, imagesCtx := errgroup.WithContext(ctx)
	group.SetLimit(10)
	for _, themeContribute := range info.ThemeContributes {
		group.Go(func() error {
			// Skip if theme path is not a json file.
			if filepath.Ext(themeContribute.Path) != ".json" {
				log.Infof("Skipping theme: %s", themeContribute.Path)
				return nil
			}

			log.Infof("Generating images for theme: %s", themeContribute.Path)
			result, err := cli.GenerateImages(imagesCtx, extensionPath, themeContribute, imagesPath)
			if err != nil {
				return fmt.Errorf("failed to generate images for %s: %w", themeContribute.Path, err)
			}

			if result != nil {
				// Override the absolute path with the relative path, which we use later to save to the database.
				result.Theme.Path = themeContribute.Path
				imagesResults = append(imagesResults, *result)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	if len(imagesResults) == 0 {
		log.Info("No images generated, skipping extension")
		return nil
	}

	// Upload images for each theme concurrency, up to a max of 10 subroutines.
	group, uploadCtx := errgroup.WithContext(ctx)
	group.SetLimit(10)

	// Generate a cache bust ID based on the job ID.
	cacheBustId := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(job.ID)).Bytes())

	slugGenerator := makeThemeSlugGenerator()
	upsertThemeWithImagesParams := make([]UpsertThemeWithImagesParams, len(imagesResults))
	for themeIndex, result := range imagesResults {
		themeSlug := slugGenerator(result.Theme.DisplayName)

		// TODO: Save themeIndex as ordinal.

		upsertThemeParams, err := convertUpsertThemeParams(themeSlug, result.Theme)
		if err != nil {
			return fmt.Errorf("failed to convert upsert theme params: %w", err)
		}

		upsertThemeWithImagesParams[themeIndex] = UpsertThemeWithImagesParams{
			Theme:  upsertThemeParams,
			Images: make([]db.UpsertImageParams, len(result.Languages)),
		}

		group.Go(func() error {
			log.Infof("Uploading images for theme: %s", result.Theme.Path)

			for languageIndex, language := range result.Languages {
				file, err := os.Open(language.SvgPath)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				defer file.Close()

				imageType := "preview"
				imageFormat := "svg"
				svgFileName := fmt.Sprintf("%s-%s-%s-%s.%s", themeSlug, language.Language.ExtName, imageType, cacheBustId, imageFormat)
				svgObjectKey := fmt.Sprintf("%s/%s", extensionSlug, svgFileName)

				log.Debugf("Uploading SVG image at %s to %s", language.SvgPath, svgObjectKey)

				_, err = w.ObjectStoreClient.PutObject(uploadCtx, &s3.PutObjectInput{
					Bucket:       aws.String(w.ObjectStoreBucket),
					Key:          aws.String(svgObjectKey),
					Body:         file,
					ContentType:  aws.String("image/svg+xml"),
					CacheControl: aws.String("public, max-age=31536000"),
				})
				if err != nil {
					return fmt.Errorf("failed to upload svg file to %s: %w", svgObjectKey, err)
				}

				svgImageUrl := fmt.Sprintf("%s/%s", w.CDNBaseUrl, svgObjectKey)
				log.Debugf("SVG image uploaded: %s", svgImageUrl)

				upsertImageParams := db.UpsertImageParams{
					Language: language.Language.ExtName,
					Type:     imageType,
					Format:   imageFormat,
					Url:      svgImageUrl,
				}

				upsertThemeWithImagesParams[themeIndex].Images[languageIndex] = upsertImageParams

				// TODO: OG image generation.
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	log.Infof("Saving extension to database")
	if err = saveExtension(ctx, w.DBPool, upsertExtensionParams, upsertThemeWithImagesParams); err != nil {
		return fmt.Errorf("failed to save extension to database: %w", err)
	}
	log.Infof("Extension saved to database")

	return nil
}

func isExtensionUpToDate(ctx context.Context, queries *db.Queries, extension marketplace.ExtensionResult) (bool, error) {
	savedExtension, err := queries.GetExtension(ctx, db.GetExtensionParams{
		ExtensionName: extension.ExtensionName,
		PublisherName: extension.Publisher.PublisherName,
		Language:      "go",
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to get extension: %w", err)
	}

	publishedAt, err := time.Parse(time.RFC3339, extension.PublishedDate)
	if err != nil {
		return false, fmt.Errorf("failed to parse publishedAt: %w", err)
	}

	return savedExtension.PublishedAt.Time.Equal(publishedAt), nil
}

func convertUpsertExtensionParams(extension marketplace.ExtensionResult) (db.UpsertExtensionParams, error) {
	params := db.UpsertExtensionParams{
		VscExtensionID:       extension.ExtensionID,
		Name:                 extension.ExtensionName,
		DisplayName:          extension.DisplayName,
		ShortDescription:     db.Text(extension.ShortDescription),
		PublisherID:          extension.Publisher.PublisherID,
		PublisherName:        extension.Publisher.PublisherName,
		PublisherDisplayName: extension.Publisher.DisplayName,
	}

	publishedAt, err := time.Parse(time.RFC3339, extension.PublishedDate)
	if err != nil {
		return params, fmt.Errorf("failed to parse publishedAt: %w", err)
	}
	params.PublishedAt = db.Timestamp(&publishedAt)

	releasedAt, err := time.Parse(time.RFC3339, extension.ReleaseDate)
	if err != nil {
		return params, fmt.Errorf("failed to parse releasedAt: %w", err)
	}
	params.ReleasedAt = db.Timestamp(&releasedAt)

	installs := findStatistic(extension.Stastistics, "install")
	params.Installs = int32(installs)

	trendingDailyStat := findStatistic(extension.Stastistics, "trendingdaily")
	trendingDaily, err := db.Numeric(&trendingDailyStat)
	if err != nil {
		return params, fmt.Errorf("failed to convert trendingDaily to numeric: %w", err)
	}
	params.TrendingDaily = trendingDaily

	trendingWeeklyStat := findStatistic(extension.Stastistics, "trendingweekly")
	trendingWeekly, err := db.Numeric(&trendingWeeklyStat)
	if err != nil {
		return params, fmt.Errorf("failed to convert trendingWeekly to numeric: %w", err)
	}
	params.TrendingWeekly = trendingWeekly

	trendingMonthlyStat := findStatistic(extension.Stastistics, "trendingmonthly")
	trendingMonthly, err := db.Numeric(&trendingMonthlyStat)
	if err != nil {
		return params, fmt.Errorf("failed to convert trendingMonthly to numeric: %w", err)
	}
	params.TrendingMonthly = trendingMonthly

	weightedRatingStat := findStatistic(extension.Stastistics, "weightedRating")
	weightedRating, err := db.Numeric(&weightedRatingStat)
	if err != nil {
		return params, fmt.Errorf("failed to convert weightedRating to numeric: %w", err)
	}
	params.WeightedRating = weightedRating

	return params, nil
}

func findStatistic(statistics []marketplace.ExtensionStatisticsResult, name string) float64 {
	for _, statistic := range statistics {
		if statistic.StatisticName == name {
			return statistic.Value
		}
	}

	return 0
}

func makeThemeSlugGenerator() func(string) string {
	themeSlugCounts := make(map[string]int)

	return func(displayName string) string {
		themeSlug := slug.MakeLang(displayName, "en")
		if count, ok := themeSlugCounts[themeSlug]; ok {
			themeSlug = fmt.Sprintf("%s-%d", themeSlug, count+1)
			themeSlugCounts[themeSlug] = count + 1
		} else {
			themeSlugCounts[themeSlug] = 1
		}

		return themeSlug
	}
}

func convertUpsertThemeParams(themeSlug string, theme cli.Theme) (db.UpsertThemeParams, error) {
	upsertThemeParams := db.UpsertThemeParams{
		Path:        theme.Path,
		DisplayName: theme.DisplayName,
		Name:        themeSlug,
	}

	editorBackground, err := colors.HexToLabString(theme.Colors.EditorBackground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert editorBackground to lab: %w", err)
	}
	upsertThemeParams.EditorBackground = editorBackground

	editorForeground, err := colors.HexToLabString(theme.Colors.EditorForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert editorForeground to lab: %w", err)
	}
	upsertThemeParams.EditorForeground = editorForeground

	activityBarBackground, err := colors.HexToLabString(theme.Colors.ActivityBarBackground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert activityBarBackground to lab: %w", err)
	}
	upsertThemeParams.ActivityBarBackground = activityBarBackground

	activityBarForeground, err := colors.HexToLabString(theme.Colors.ActivityBarForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert activityBarForeground to lab: %w", err)
	}
	upsertThemeParams.ActivityBarForeground = activityBarForeground

	activityBarInActiveForeground, err := colors.HexToLabString(theme.Colors.ActivityBarInActiveForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert activityBarInActiveForeground to lab: %w", err)
	}
	upsertThemeParams.ActivityBarInActiveForeground = activityBarInActiveForeground

	if theme.Colors.ActivityBarBorder != nil {
		activityBarBorder, err := colors.HexToLabString(*theme.Colors.ActivityBarBorder)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert activityBarBorder to lab: %w", err)
		}
		upsertThemeParams.ActivityBarBorder = &activityBarBorder
	}

	activityBarActiveBorder, err := colors.HexToLabString(theme.Colors.ActivityBarActiveBorder)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert activityBarActiveBorder to lab: %w", err)
	}
	upsertThemeParams.ActivityBarActiveBorder = activityBarActiveBorder

	if theme.Colors.ActivityBarActiveBackground != nil {
		activityBarActiveBackground, err := colors.HexToLabString(*theme.Colors.ActivityBarActiveBackground)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert activityBarActiveBackground to lab: %w", err)
		}
		upsertThemeParams.ActivityBarActiveBackground = &activityBarActiveBackground
	}

	activityBarBadgeBackground, err := colors.HexToLabString(theme.Colors.ActivityBarBadgeBackground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert activityBarBadgeBackground to lab: %w", err)
	}
	upsertThemeParams.ActivityBarBadgeBackground = activityBarBadgeBackground

	activityBarBadgeForeground, err := colors.HexToLabString(theme.Colors.ActivityBarBadgeForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert activityBarBadgeForeground to lab: %w", err)
	}
	upsertThemeParams.ActivityBarBadgeForeground = activityBarBadgeForeground

	if theme.Colors.TabsContainerBackground != nil {
		tabsContainerBackground, err := colors.HexToLabString(*theme.Colors.TabsContainerBackground)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert tabsContainerBackground to lab: %w", err)
		}
		upsertThemeParams.TabsContainerBackground = &tabsContainerBackground
	}

	if theme.Colors.TabsContainerBorder != nil {
		tabsContainerBorder, err := colors.HexToLabString(*theme.Colors.TabsContainerBorder)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert tabsContainerBorder to lab: %w", err)
		}
		upsertThemeParams.TabsContainerBorder = &tabsContainerBorder
	}

	if theme.Colors.StatusBarBackground != nil {
		statusBarBackground, err := colors.HexToLabString(*theme.Colors.StatusBarBackground)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert statusBarBackground to lab: %w", err)
		}
		upsertThemeParams.StatusBarBackground = &statusBarBackground
	}

	statusBarForeground, err := colors.HexToLabString(theme.Colors.StatusBarForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert statusBarForeground to lab: %w", err)
	}
	upsertThemeParams.StatusBarForeground = statusBarForeground

	if theme.Colors.StatusBarBorder != nil {
		statusBarBorder, err := colors.HexToLabString(*theme.Colors.StatusBarBorder)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert statusBarBorder to lab: %w", err)
		}
		upsertThemeParams.StatusBarBorder = &statusBarBorder
	}

	if theme.Colors.TabActiveBackground != nil {
		tabActiveBackground, err := colors.HexToLabString(*theme.Colors.TabActiveBackground)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert tabActiveBackground to lab: %w", err)
		}
		upsertThemeParams.TabActiveBackground = &tabActiveBackground
	}

	if theme.Colors.TabInactiveBackground != nil {
		tabInactiveBackground, err := colors.HexToLabString(*theme.Colors.TabInactiveBackground)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert tabInactiveBackground to lab: %w", err)
		}
		upsertThemeParams.TabInactiveBackground = &tabInactiveBackground
	}

	tabActiveForeground, err := colors.HexToLabString(theme.Colors.TabActiveForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert tabActiveForeground to lab: %w", err)
	}
	upsertThemeParams.TabActiveForeground = tabActiveForeground

	tabBorder, err := colors.HexToLabString(theme.Colors.TabBorder)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert tabBorder to lab: %w", err)
	}
	upsertThemeParams.TabBorder = tabBorder

	if theme.Colors.TabActiveBorder != nil {
		tabActiveBorder, err := colors.HexToLabString(*theme.Colors.TabActiveBorder)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert tabActiveBorder to lab: %w", err)
		}
		upsertThemeParams.TabActiveBorder = &tabActiveBorder
	}

	if theme.Colors.TabActiveBorderTop != nil {
		tabActiveBorderTop, err := colors.HexToLabString(*theme.Colors.TabActiveBorderTop)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert tabActiveBorderTop to lab: %w", err)
		}
		upsertThemeParams.TabActiveBorderTop = &tabActiveBorderTop
	}

	titleBarActiveBackground, err := colors.HexToLabString(theme.Colors.TitleBarActiveBackground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert titleBarActiveBackground to lab: %w", err)
	}
	upsertThemeParams.TitleBarActiveBackground = titleBarActiveBackground

	titleBarActiveForeground, err := colors.HexToLabString(theme.Colors.TitleBarActiveForeground)
	if err != nil {
		return upsertThemeParams, fmt.Errorf("failed to convert titleBarActiveForeground to lab: %w", err)
	}
	upsertThemeParams.TitleBarActiveForeground = titleBarActiveForeground

	if theme.Colors.TitleBarBorder != nil {
		titleBarBorder, err := colors.HexToLabString(*theme.Colors.TitleBarBorder)
		if err != nil {
			return upsertThemeParams, fmt.Errorf("failed to convert titleBarBorder to lab: %w", err)
		}
		upsertThemeParams.TitleBarBorder = &titleBarBorder
	}

	return upsertThemeParams, nil
}

type UpsertThemeWithImagesParams struct {
	Theme  db.UpsertThemeParams
	Images []db.UpsertImageParams
}

func saveExtension(ctx context.Context, dbPool *pgxpool.Pool, extension db.UpsertExtensionParams, themes []UpsertThemeWithImagesParams) error {
	return pgx.BeginFunc(ctx, dbPool, func(tx pgx.Tx) error {
		queries := db.New(tx)

		// TODO: Check if an extension exists with the same name and publisher name but with
		// a different vsc extension id. Delete the older one if it's the same publisher?
		// Otherwise skip the extension if the publisher is different instead of erroring.

		// Upsert extension.
		extension, err := queries.UpsertExtension(ctx, extension)
		if err != nil {
			return fmt.Errorf("failed to upsert extension: %w", err)
		}

		// Upsert themes and images.
		for _, themeWithImages := range themes {
			// Set extension ID for each theme.
			themeWithImages.Theme.ExtensionID = extension.ID

			// Upsert theme.
			theme, err := queries.UpsertTheme(ctx, themeWithImages.Theme)
			if err != nil {
				return fmt.Errorf("failed to upsert theme: %w", err)
			}

			// Upsert images.
			for _, image := range themeWithImages.Images {
				// Set theme ID for each image.
				image.ThemeID = theme.ID

				// Upsert image.
				if _, err := queries.UpsertImage(ctx, image); err != nil {
					return fmt.Errorf("failed to upsert image: %w", err)
				}
			}
		}

		// TODO: Delete old themes and images.

		return nil
	})
}
