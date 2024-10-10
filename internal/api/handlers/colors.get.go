package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/middleware"
	"github.com/vscodethemes/backend/internal/colors"
	"github.com/vscodethemes/backend/internal/db"
)

var GetColorsOperation = huma.Operation{
	OperationID: "get-colors",
	Method:      http.MethodGet,
	Path:        "/themes/colors",
	Summary:     "Get Colors",
	Description: "Count the number of colors for all themes",
	Tags:        []string{"Colors"},
	Errors:      []int{http.StatusNotFound},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("colors:read"),
	},
}

type GetColorsInput struct {
	Anchor int `query:"anchor" default:"10" example:"10" doc:"The closest multiple to round to"`
}

type GetColorsOutput struct {
	Body struct {
		Colors []Color `json:"colors"`
	}
}

type Color struct {
	Hex   string `json:"hex"`
	Count int64  `json:"count"`
}

func (h Handler) GetColors(ctx context.Context, input *GetColorsInput) (*GetColorsOutput, error) {
	queries := db.New(h.DBPool)

	rows, err := queries.GetColorCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get color counts: %w", err)
	}

	colorMap := map[string]*Color{}
	for _, row := range rows {
		x, y, z, err := colors.LabStringToXyz(row.Color)
		if err != nil {
			h.Logger.Error("failed to convert lab to xyz", "color", row.Color, "error", err)
			continue
		}

		xR := roundToAnchor(x, input.Anchor)
		yR := roundToAnchor(y, input.Anchor)
		zR := roundToAnchor(z, input.Anchor)

		colorKey := fmt.Sprintf("%d,%d,%d", xR, yR, zR)
		colorMapValue, ok := colorMap[colorKey]
		if ok {
			colorMapValue.Count += row.Count
		} else {
			r, g, b := colors.XyzToRgb(x, y, z)

			colorMap[colorKey] = &Color{
				Hex:   colors.RgbToHex(r, g, b),
				Count: row.Count,
			}
		}
	}

	resp := &GetColorsOutput{}
	resp.Body.Colors = []Color{}
	for _, colorValue := range colorMap {
		resp.Body.Colors = append(resp.Body.Colors, *colorValue)
	}

	// Sort by count descending.
	sort.Slice(resp.Body.Colors, func(i, j int) bool {
		return resp.Body.Colors[i].Count > resp.Body.Colors[j].Count
	})

	return resp, nil
}

func roundToAnchor(x float64, anchor int) int {
	if anchor == 0 {
		return int(x)
	}
	return anchor * int(math.Round(x/float64(anchor)))
}
