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

DROP INDEX IF EXISTS public.idx_notifications_user_id_created_at;
DROP INDEX IF EXISTS public.idx_chats_messages_sender_receiver;
DROP INDEX IF EXISTS public.idx_chats_messages_ride;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.notifications DROP CONSTRAINT IF EXISTS notifications_pkey;
ALTER TABLE IF EXISTS ONLY public.device_tokens DROP CONSTRAINT IF EXISTS device_tokens_pkey;
ALTER TABLE IF EXISTS ONLY public.chats_messages DROP CONSTRAINT IF EXISTS chats_messages_pkey;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.notifications;
DROP TABLE IF EXISTS public.device_tokens;
DROP TABLE IF EXISTS public.chats_messages;
DROP EXTENSION IF EXISTS "uuid-ossp";
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
-- Name: chats_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chats_messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    sender_id uuid NOT NULL,
    receiver_id uuid NOT NULL,
    ride_id uuid,
    content text NOT NULL,
    status character varying(50) DEFAULT 'sent'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    reply_to_id uuid
);


--
-- Name: device_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.device_tokens (
    user_id uuid NOT NULL,
    fcm_token character varying(255) NOT NULL,
    platform character varying(50),
    language character varying(10) DEFAULT 'ru'::character varying,
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    title character varying(255) NOT NULL,
    body text NOT NULL,
    type character varying(50) NOT NULL,
    is_read boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
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
-- Data for Name: chats_messages; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.chats_messages (id, sender_id, receiver_id, ride_id, content, status, created_at, updated_at, deleted_at, reply_to_id) FROM stdin;
84b47ed9-bf3b-4617-82f3-742b5bde4977	60513e3b-5f2a-4b2c-a37b-6e726c089d00	60513e3b-5f2a-4b2c-a37b-6e726c089d00	\N	test-reply-probe	sent	2026-08-20 18:21:15.149299+00	2026-08-20 18:21:15.149299+00	\N	11111111-1111-1111-1111-111111111111
bc7fc6dc-58b5-4ca0-a1c3-e9e39f2f4a03	60513e3b-5f2a-4b2c-a37b-6e726c089d00	60513e3b-5f2a-4b2c-a37b-6e726c089d00	\N	probe-plain-1787244586732	sent	2026-08-20 18:21:15.155706+00	2026-08-20 18:21:15.155706+00	\N	\N
0fa8199e-dbc1-4958-b82d-b94251550674	60513e3b-5f2a-4b2c-a37b-6e726c089d00	60513e3b-5f2a-4b2c-a37b-6e726c089d00	\N	probe-1787345588060	sent	2026-08-21 20:53:08.129116+00	2026-08-21 20:53:08.129116+00	\N	11111111-1111-1111-1111-111111111111
\.


--
-- Data for Name: device_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.device_tokens (user_id, fcm_token, platform, language, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.notifications (id, user_id, title, body, type, is_read, created_at, updated_at, deleted_at) FROM stdin;
dc647c41-efa9-4782-ae0f-fcf65c323401	8bc95322-9fc9-4903-b893-79c37aa115bc	Предложение отклонено	Водитель не ответил на ваше предложение цены в течение 30 минут. Бронирование отклонено.	booking.rejected	t	2026-08-19 02:39:46.487423+00	2026-08-19 12:45:15.523685+00	\N
ccfd5dd9-48b9-412a-9421-d2986d4cf14e	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	System Notification	New price negotiation requested by passenger!	negotiation_requested	t	2026-08-19 02:08:48.45414+00	2026-08-19 17:41:27.105793+00	\N
0621c351-4cd2-4c67-bffa-7c60c074dc30	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	System Notification	You have a new message!	subscription_expired	t	2026-08-20 00:49:36.388091+00	2026-08-20 00:49:41.089994+00	\N
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Name: chats_messages chats_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chats_messages
    ADD CONSTRAINT chats_messages_pkey PRIMARY KEY (id);


--
-- Name: device_tokens device_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.device_tokens
    ADD CONSTRAINT device_tokens_pkey PRIMARY KEY (user_id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_chats_messages_ride; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chats_messages_ride ON public.chats_messages USING btree (ride_id);


--
-- Name: idx_chats_messages_sender_receiver; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chats_messages_sender_receiver ON public.chats_messages USING btree (sender_id, receiver_id);


--
-- Name: idx_notifications_user_id_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_id_created_at ON public.notifications USING btree (user_id, created_at DESC);


--
-- PostgreSQL database dump complete
--

