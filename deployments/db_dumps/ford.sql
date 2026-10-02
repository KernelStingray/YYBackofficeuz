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

DROP INDEX IF EXISTS public.idx_fraud_events_user_id;
ALTER TABLE IF EXISTS ONLY public.trust_scores DROP CONSTRAINT IF EXISTS trust_scores_pkey;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.fraud_events DROP CONSTRAINT IF EXISTS fraud_events_pkey;
DROP TABLE IF EXISTS public.trust_scores;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.fraud_events;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: fraud_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fraud_events (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    event_type character varying(50) NOT NULL,
    description text,
    penalty_applied numeric(5,2) DEFAULT 0.00 NOT NULL,
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
-- Name: trust_scores; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trust_scores (
    user_id uuid NOT NULL,
    score numeric(5,2) DEFAULT 100.00 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Data for Name: fraud_events; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fraud_events (id, user_id, event_type, description, penalty_applied, created_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: trust_scores; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.trust_scores (user_id, score, updated_at, deleted_at) FROM stdin;
\.


--
-- Name: fraud_events fraud_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fraud_events
    ADD CONSTRAINT fraud_events_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: trust_scores trust_scores_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trust_scores
    ADD CONSTRAINT trust_scores_pkey PRIMARY KEY (user_id);


--
-- Name: idx_fraud_events_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fraud_events_user_id ON public.fraud_events USING btree (user_id);


--
-- PostgreSQL database dump complete
--

