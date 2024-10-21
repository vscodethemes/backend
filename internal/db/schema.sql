SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: cube; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS cube WITH SCHEMA public;


--
-- Name: EXTENSION cube; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION cube IS 'data type for multidimensional cubes';


--
-- Name: river_job_state; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.river_job_state AS ENUM (
    'available',
    'cancelled',
    'completed',
    'discarded',
    'pending',
    'retryable',
    'running',
    'scheduled'
);


--
-- Name: river_job_state_in_bitmask(bit, public.river_job_state); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.river_job_state_in_bitmask(bitmask bit, state public.river_job_state) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT CASE state
        WHEN 'available' THEN get_bit(bitmask, 7)
        WHEN 'cancelled' THEN get_bit(bitmask, 6)
        WHEN 'completed' THEN get_bit(bitmask, 5)
        WHEN 'discarded' THEN get_bit(bitmask, 4)
        WHEN 'pending' THEN get_bit(bitmask, 3)
        WHEN 'retryable' THEN get_bit(bitmask, 2)
        WHEN 'running' THEN get_bit(bitmask, 1)
        WHEN 'scheduled' THEN get_bit(bitmask, 0)
        ELSE 0
    END = 1;
$$;


--
-- Name: tsv_trigger(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.tsv_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
	  SELECT
		  setweight(to_tsvector('english', coalesce(new.name,'')), 'A') ||
      setweight(to_tsvector('english', coalesce(e.name,'')), 'A') ||
      setweight(to_tsvector('english', coalesce(e.publisher_name,'')), 'A') ||
      setweight(to_tsvector('english', coalesce(new.display_name,'')), 'B') ||
      setweight(to_tsvector('english', coalesce(e.display_name,'')), 'C') ||
      setweight(to_tsvector('english', coalesce(e.publisher_display_name,'')), 'C') ||
      setweight(to_tsvector('english', coalesce(e.short_description,'')), 'D')
	  INTO new.tsv
	  FROM extensions e
	  WHERE e.id = new.extension_id;

	  RETURN new;
END
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: extensions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.extensions (
    id bigint NOT NULL,
    vsc_extension_id text NOT NULL,
    name text NOT NULL,
    display_name text NOT NULL,
    short_description text,
    publisher_id text NOT NULL,
    publisher_name text NOT NULL,
    publisher_display_name text NOT NULL,
    installs integer NOT NULL,
    trending_daily numeric NOT NULL,
    trending_weekly numeric NOT NULL,
    trending_monthly numeric NOT NULL,
    weighted_rating numeric NOT NULL,
    published_at timestamp without time zone NOT NULL,
    released_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: extensions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.extensions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: extensions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.extensions_id_seq OWNED BY public.extensions.id;


--
-- Name: images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.images (
    id bigint NOT NULL,
    theme_id bigint NOT NULL,
    language text NOT NULL,
    type text NOT NULL,
    format text NOT NULL,
    url text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: images_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.images_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: images_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.images_id_seq OWNED BY public.images.id;


--
-- Name: river_client; Type: TABLE; Schema: public; Owner: -
--

CREATE UNLOGGED TABLE public.river_client (
    id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    paused_at timestamp with time zone,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT name_length CHECK (((char_length(id) > 0) AND (char_length(id) < 128)))
);


--
-- Name: river_client_queue; Type: TABLE; Schema: public; Owner: -
--

CREATE UNLOGGED TABLE public.river_client_queue (
    river_client_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    max_workers bigint DEFAULT 0 NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    num_jobs_completed bigint DEFAULT 0 NOT NULL,
    num_jobs_running bigint DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT name_length CHECK (((char_length(name) > 0) AND (char_length(name) < 128))),
    CONSTRAINT num_jobs_completed_zero_or_positive CHECK ((num_jobs_completed >= 0)),
    CONSTRAINT num_jobs_running_zero_or_positive CHECK ((num_jobs_running >= 0))
);


--
-- Name: river_job; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.river_job (
    id bigint NOT NULL,
    state public.river_job_state DEFAULT 'available'::public.river_job_state NOT NULL,
    attempt smallint DEFAULT 0 NOT NULL,
    max_attempts smallint NOT NULL,
    attempted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    finalized_at timestamp with time zone,
    scheduled_at timestamp with time zone DEFAULT now() NOT NULL,
    priority smallint DEFAULT 1 NOT NULL,
    args jsonb NOT NULL,
    attempted_by text[],
    errors jsonb[],
    kind text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    queue text DEFAULT 'default'::text NOT NULL,
    tags character varying(255)[] DEFAULT '{}'::character varying[] NOT NULL,
    unique_key bytea,
    unique_states bit(8),
    CONSTRAINT finalized_or_finalized_at_null CHECK ((((finalized_at IS NULL) AND (state <> ALL (ARRAY['cancelled'::public.river_job_state, 'completed'::public.river_job_state, 'discarded'::public.river_job_state]))) OR ((finalized_at IS NOT NULL) AND (state = ANY (ARRAY['cancelled'::public.river_job_state, 'completed'::public.river_job_state, 'discarded'::public.river_job_state]))))),
    CONSTRAINT kind_length CHECK (((char_length(kind) > 0) AND (char_length(kind) < 128))),
    CONSTRAINT max_attempts_is_positive CHECK ((max_attempts > 0)),
    CONSTRAINT priority_in_range CHECK (((priority >= 1) AND (priority <= 4))),
    CONSTRAINT queue_length CHECK (((char_length(queue) > 0) AND (char_length(queue) < 128)))
);


--
-- Name: river_job_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.river_job_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: river_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.river_job_id_seq OWNED BY public.river_job.id;


--
-- Name: river_leader; Type: TABLE; Schema: public; Owner: -
--

CREATE UNLOGGED TABLE public.river_leader (
    elected_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    leader_id text NOT NULL,
    name text DEFAULT 'default'::text NOT NULL,
    CONSTRAINT leader_id_length CHECK (((char_length(leader_id) > 0) AND (char_length(leader_id) < 128))),
    CONSTRAINT name_length CHECK ((name = 'default'::text))
);


--
-- Name: river_queue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.river_queue (
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    paused_at timestamp with time zone,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL
);


--
-- Name: themes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.themes (
    id bigint NOT NULL,
    extension_id bigint NOT NULL,
    path text NOT NULL,
    name text NOT NULL,
    display_name text NOT NULL,
    editor_background public.cube NOT NULL,
    editor_foreground public.cube NOT NULL,
    activity_bar_background public.cube NOT NULL,
    activity_bar_foreground public.cube NOT NULL,
    activity_bar_in_active_foreground public.cube NOT NULL,
    activity_bar_border public.cube,
    activity_bar_active_border public.cube NOT NULL,
    activity_bar_active_background public.cube,
    activity_bar_badge_background public.cube NOT NULL,
    activity_bar_badge_foreground public.cube NOT NULL,
    tabs_container_background public.cube,
    tabs_container_border public.cube,
    status_bar_background public.cube,
    status_bar_foreground public.cube NOT NULL,
    status_bar_border public.cube,
    tab_active_background public.cube,
    tab_inactive_background public.cube,
    tab_active_foreground public.cube NOT NULL,
    tab_border public.cube NOT NULL,
    tab_active_border public.cube,
    tab_active_border_top public.cube,
    title_bar_active_background public.cube NOT NULL,
    title_bar_active_foreground public.cube NOT NULL,
    title_bar_border public.cube,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    tsv tsvector NOT NULL
);


--
-- Name: themes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.themes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: themes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.themes_id_seq OWNED BY public.themes.id;


--
-- Name: extensions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extensions ALTER COLUMN id SET DEFAULT nextval('public.extensions_id_seq'::regclass);


--
-- Name: images id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images ALTER COLUMN id SET DEFAULT nextval('public.images_id_seq'::regclass);


--
-- Name: river_job id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_job ALTER COLUMN id SET DEFAULT nextval('public.river_job_id_seq'::regclass);


--
-- Name: themes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.themes ALTER COLUMN id SET DEFAULT nextval('public.themes_id_seq'::regclass);


--
-- Name: extensions extensions_name_publisher_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extensions
    ADD CONSTRAINT extensions_name_publisher_name_key UNIQUE (name, publisher_name);


--
-- Name: extensions extensions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extensions
    ADD CONSTRAINT extensions_pkey PRIMARY KEY (id);


--
-- Name: extensions extensions_vsc_extension_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extensions
    ADD CONSTRAINT extensions_vsc_extension_id_key UNIQUE (vsc_extension_id);


--
-- Name: images images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_pkey PRIMARY KEY (id);


--
-- Name: images images_theme_id_language_type_format_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_theme_id_language_type_format_key UNIQUE (theme_id, language, type, format);


--
-- Name: river_client river_client_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_client
    ADD CONSTRAINT river_client_pkey PRIMARY KEY (id);


--
-- Name: river_client_queue river_client_queue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_client_queue
    ADD CONSTRAINT river_client_queue_pkey PRIMARY KEY (river_client_id, name);


--
-- Name: river_job river_job_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_job
    ADD CONSTRAINT river_job_pkey PRIMARY KEY (id);


--
-- Name: river_leader river_leader_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_leader
    ADD CONSTRAINT river_leader_pkey PRIMARY KEY (name);


--
-- Name: river_queue river_queue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_queue
    ADD CONSTRAINT river_queue_pkey PRIMARY KEY (name);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: themes themes_extension_id_path_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.themes
    ADD CONSTRAINT themes_extension_id_path_key UNIQUE (extension_id, path);


--
-- Name: themes themes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.themes
    ADD CONSTRAINT themes_pkey PRIMARY KEY (id);


--
-- Name: river_job_args_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX river_job_args_index ON public.river_job USING gin (args);


--
-- Name: river_job_kind; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX river_job_kind ON public.river_job USING btree (kind);


--
-- Name: river_job_metadata_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX river_job_metadata_index ON public.river_job USING gin (metadata);


--
-- Name: river_job_prioritized_fetching_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX river_job_prioritized_fetching_index ON public.river_job USING btree (state, queue, priority, scheduled_at, id);


--
-- Name: river_job_state_and_finalized_at_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX river_job_state_and_finalized_at_index ON public.river_job USING btree (state, finalized_at) WHERE (finalized_at IS NOT NULL);


--
-- Name: river_job_unique_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX river_job_unique_idx ON public.river_job USING btree (unique_key) WHERE ((unique_key IS NOT NULL) AND (unique_states IS NOT NULL) AND public.river_job_state_in_bitmask(unique_states, state));


--
-- Name: themes_tsv_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX themes_tsv_idx ON public.themes USING gist (tsv);


--
-- Name: themes tsvupdate; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER tsvupdate BEFORE INSERT OR UPDATE ON public.themes FOR EACH ROW EXECUTE FUNCTION public.tsv_trigger();


--
-- Name: images images_theme_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.images
    ADD CONSTRAINT images_theme_id_fkey FOREIGN KEY (theme_id) REFERENCES public.themes(id) ON DELETE CASCADE;


--
-- Name: river_client_queue river_client_queue_river_client_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.river_client_queue
    ADD CONSTRAINT river_client_queue_river_client_id_fkey FOREIGN KEY (river_client_id) REFERENCES public.river_client(id) ON DELETE CASCADE;


--
-- Name: themes themes_extension_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.themes
    ADD CONSTRAINT themes_extension_id_fkey FOREIGN KEY (extension_id) REFERENCES public.extensions(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20240820234134'),
    ('20240922015622'),
    ('20240930011343'),
    ('20241021160435');
