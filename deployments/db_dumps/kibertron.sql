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

DROP INDEX IF EXISTS public.idx_audit_logs_created_at;
DROP INDEX IF EXISTS public.idx_audit_logs_admin_id;
DROP INDEX IF EXISTS public.idx_audit_logs_action;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.remote_configs DROP CONSTRAINT IF EXISTS remote_configs_pkey;
ALTER TABLE IF EXISTS ONLY public.audit_logs DROP CONSTRAINT IF EXISTS audit_logs_pkey;
ALTER TABLE IF EXISTS ONLY public.admins DROP CONSTRAINT IF EXISTS admins_pkey;
ALTER TABLE IF EXISTS ONLY public.admins DROP CONSTRAINT IF EXISTS admins_login_key;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.remote_configs;
DROP TABLE IF EXISTS public.audit_logs;
DROP TABLE IF EXISTS public.admins;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admins (
    id uuid NOT NULL,
    login character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    role character varying(50) DEFAULT 'moderator'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    admin_id uuid NOT NULL,
    action character varying(10) NOT NULL,
    entity character varying(255) NOT NULL,
    target_id character varying(255),
    payload jsonb,
    ip_address character varying(45),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: remote_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.remote_configs (
    key character varying(255) NOT NULL,
    value jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
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
-- Data for Name: admins; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.admins (id, login, password_hash, role, created_at, deleted_at) FROM stdin;
4ddc9e6c-9ca9-4d65-bb73-94c379844f33	KernelStingray	$2a$10$swDRSsphpmQsnzJi7N29RO6lWV5YcQhfhMp.korveK7Ts9iZt7Bf2	admin	2026-08-18 01:03:30.339828+00	\N
\.


--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.audit_logs (id, admin_id, action, entity, target_id, payload, ip_address, created_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: remote_configs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.remote_configs (key, value, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Name: admins admins_login_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_login_key UNIQUE (login);


--
-- Name: admins admins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admins
    ADD CONSTRAINT admins_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: remote_configs remote_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.remote_configs
    ADD CONSTRAINT remote_configs_pkey PRIMARY KEY (key);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_action ON public.audit_logs USING btree (action);


--
-- Name: idx_audit_logs_admin_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_admin_id ON public.audit_logs USING btree (admin_id);


--
-- Name: idx_audit_logs_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_created_at ON public.audit_logs USING btree (created_at);


--
-- PostgreSQL database dump complete
--

