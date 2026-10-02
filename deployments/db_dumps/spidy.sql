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

DROP INDEX IF EXISTS public.idx_tickets_user_id;
ALTER TABLE IF EXISTS ONLY public.tickets DROP CONSTRAINT IF EXISTS tickets_pkey;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
DROP TABLE IF EXISTS public.tickets;
DROP TABLE IF EXISTS public.schema_migrations;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: tickets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tickets (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    subject character varying(255) NOT NULL,
    description text NOT NULL,
    status character varying(50) DEFAULT 'OPEN'::character varying NOT NULL,
    resolution text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: tickets; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.tickets (id, user_id, subject, description, status, resolution, created_at, updated_at, deleted_at) FROM stdin;
aa1385ae-d602-4e33-b850-bf4116268e9b	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	Жалоба	Ey masha\n\n--- texnik ma'lumot ---\nIlova: 1.0.0+1014 (android)\nQurilma: 38feaeb3-aae2-4bc6-a8b6-e4e8014c71c6	OPEN	\N	2026-08-18 18:37:02.119072+00	2026-08-18 18:37:02.119072+00	\N
\.


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: tickets tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_pkey PRIMARY KEY (id);


--
-- Name: idx_tickets_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tickets_user_id ON public.tickets USING btree (user_id);


--
-- PostgreSQL database dump complete
--

