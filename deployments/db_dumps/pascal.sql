--
-- PostgreSQL database dump
--

-- Dumped from database version 15.4
-- Dumped by pg_dump version 15.4

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

DROP INDEX IF EXISTS public.idx_driver_location_history_geom;
DROP INDEX IF EXISTS public.idx_driver_location_history_driver_id;
DROP INDEX IF EXISTS public.idx_driver_location_history_created_at;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.driver_location_history DROP CONSTRAINT IF EXISTS driver_location_history_pkey;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.driver_location_history;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS postgis;
--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry and geography spatial types and functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: driver_location_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.driver_location_history (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    driver_id uuid NOT NULL,
    geom public.geometry(Point,4326) NOT NULL,
    speed numeric DEFAULT 0,
    heading numeric DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Data for Name: driver_location_history; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.driver_location_history (id, driver_id, geom, speed, heading, created_at, deleted_at) FROM stdin;
e6b445db-9158-463b-bdf8-af09b7aa43aa	60513e3b-5f2a-4b2c-a37b-6e726c089d00	0101000020E610000043AA285E654F514018B49080D1A74440	\N	\N	2026-08-20 18:44:57.355103+00	\N
42044d9b-926d-495b-abe8-80d85571a5a8	60513e3b-5f2a-4b2c-a37b-6e726c089d00	0101000020E610000043AA285E654F514018B49080D1A74440	\N	\N	2026-08-21 20:51:57.321622+00	\N
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: spatial_ref_sys; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.spatial_ref_sys (srid, auth_name, auth_srid, srtext, proj4text) FROM stdin;
\.


--
-- Name: driver_location_history driver_location_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.driver_location_history
    ADD CONSTRAINT driver_location_history_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_driver_location_history_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_driver_location_history_created_at ON public.driver_location_history USING btree (created_at);


--
-- Name: idx_driver_location_history_driver_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_driver_location_history_driver_id ON public.driver_location_history USING btree (driver_id);


--
-- Name: idx_driver_location_history_geom; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_driver_location_history_geom ON public.driver_location_history USING gist (geom);


--
-- PostgreSQL database dump complete
--

