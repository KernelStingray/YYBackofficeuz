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

ALTER TABLE IF EXISTS ONLY public.user_documents DROP CONSTRAINT IF EXISTS user_documents_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.reviews DROP CONSTRAINT IF EXISTS reviews_reviewer_id_fkey;
ALTER TABLE IF EXISTS ONLY public.reviews DROP CONSTRAINT IF EXISTS reviews_reviewee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.cars DROP CONSTRAINT IF EXISTS cars_user_id_fkey;
DROP INDEX IF EXISTS public.idx_users_phone;
DROP INDEX IF EXISTS public.idx_user_documents_user_id;
DROP INDEX IF EXISTS public.idx_reviews_users;
DROP INDEX IF EXISTS public.idx_cars_user_id;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_phone_number_key;
ALTER TABLE IF EXISTS ONLY public.user_documents DROP CONSTRAINT IF EXISTS user_documents_pkey;
ALTER TABLE IF EXISTS ONLY public.user_blocks DROP CONSTRAINT IF EXISTS user_blocks_pkey;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.reviews DROP CONSTRAINT IF EXISTS reviews_pkey;
ALTER TABLE IF EXISTS ONLY public.reports DROP CONSTRAINT IF EXISTS reports_pkey;
ALTER TABLE IF EXISTS ONLY public.cars DROP CONSTRAINT IF EXISTS cars_pkey;
ALTER TABLE IF EXISTS ONLY public.cars DROP CONSTRAINT IF EXISTS cars_gov_number_key;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.user_documents;
DROP TABLE IF EXISTS public.user_blocks;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.reviews;
DROP TABLE IF EXISTS public.reports;
DROP TABLE IF EXISTS public.cars;
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
-- Name: cars; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cars (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    gov_number character varying(20) NOT NULL,
    brand character varying(100) NOT NULL,
    model character varying(100) NOT NULL,
    tech_passport_photo character varying(500),
    driver_license_photo character varying(500),
    tech_passport_front character varying(500),
    tech_passport_back character varying(500),
    driver_license_front character varying(500),
    driver_license_back character varying(500),
    car_front_photo character varying(500),
    car_back_photo character varying(500),
    car_side_photo character varying(500),
    moderation_status character varying(50) DEFAULT 'pending'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp with time zone
);


--
-- Name: reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reports (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    reporter_id uuid NOT NULL,
    reported_id uuid NOT NULL,
    reason text NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    reviewer_id uuid NOT NULL,
    reviewee_id uuid NOT NULL,
    ride_id uuid NOT NULL,
    rating integer NOT NULL,
    comment text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp with time zone,
    CONSTRAINT reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: user_blocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_blocks (
    blocker_id uuid NOT NULL,
    blocked_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: user_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_documents (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    document_type character varying(50) NOT NULL,
    file_url character varying(500) NOT NULL,
    verification_status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp with time zone
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    gender character varying(20),
    phone_number character varying(20) NOT NULL,
    password_hash character varying(255),
    role character varying(20) DEFAULT 'passenger'::character varying NOT NULL,
    kyc_status character varying(20) DEFAULT 'none'::character varying,
    kyc_rejection_reason character varying(500),
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    locale character varying(10) DEFAULT 'ru'::character varying,
    avatar_url character varying(500),
    rating double precision DEFAULT 5.0 NOT NULL,
    trusted_devices text[] DEFAULT '{}'::text[],
    rating_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp with time zone
);


--
-- Data for Name: cars; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.cars (id, user_id, gov_number, brand, model, tech_passport_photo, driver_license_photo, moderation_status, created_at, updated_at, deleted_at) FROM stdin;
d336419a-6ef2-4fdc-b898-f141e295d4b2	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	50A123SS	ravon	porsche	https://api.yolyolakay.xyz/yolyolakay-media/3cce0970-5f57-4d72-b483-0bd190f1ec24.jpg	\N	pending	2026-08-20 00:35:04.80483+00	2026-08-20 00:35:04.80483+00	\N
d7e07d74-1b01-457d-ad81-8be59f9a7c18	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	01W333WR	ravon	ravon spark	https://api.yolyolakay.xyz/yolyolakay-media/e0695a7b-ca70-4794-bc1f-2216865ecb6d.jpg	\N	pending	2026-08-20 00:36:39.573057+00	2026-08-20 00:36:39.573057+00	\N
c6ced4f1-1fb7-4b41-851a-43b653c4d7bd	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	01R125MK	Chevrolet	Malibu	https://api.yolyolakay.xyz/yolyolakay-media/a68186f8-1219-478a-9942-86de66d959a7.png	\N	approved	2026-08-18 01:12:44.263752+00	2026-08-21 13:08:04.196135+00	\N
8de33592-4328-49cf-ba55-bce8898af5d1	60513e3b-5f2a-4b2c-a37b-6e726c089d00	01A999AA	Chevrolet	Cobalt	\N	\N	pending	2026-08-21 20:51:56.945182+00	2026-08-21 20:51:56.945182+00	\N
\.


--
-- Data for Name: reports; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reports (id, reporter_id, reported_id, reason, created_at) FROM stdin;
3afcd583-02e8-4236-a4a5-01befd025928	8bc95322-9fc9-4903-b893-79c37aa115bc	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	fraud	2026-08-19 02:55:32.576074+00
d494b179-c20d-4289-84fa-649bfc733b7b	8bc95322-9fc9-4903-b893-79c37aa115bc	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	other	2026-08-20 04:43:48.503545+00
\.


--
-- Data for Name: reviews; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reviews (id, reviewer_id, reviewee_id, ride_id, rating, comment, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: user_blocks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_blocks (blocker_id, blocked_id, created_at) FROM stdin;
8bc95322-9fc9-4903-b893-79c37aa115bc	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	2026-08-19 02:55:23.155225+00
\.


--
-- Data for Name: user_documents; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_documents (id, user_id, document_type, file_url, verification_status, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, first_name, last_name, gender, phone_number, password_hash, role, kyc_status, status, locale, avatar_url, rating, trusted_devices, rating_count, created_at, updated_at, deleted_at) FROM stdin;
44b88e16-3feb-4bce-850f-0c2d4b47ff80	Комил	Каримов	male	+998998996323	\N	passenger	none	active	ru	\N	5	{4d3c3eb3-3e0b-4145-98a1-e384518a973d}	0	2026-08-19 15:35:14.575773+00	2026-08-19 15:35:29.720539+00	\N
da5a196d-c70d-4e0b-b0ea-e4133730f0cf	Ulugbek	Raximov	male	+998919645859	\N	passenger	none	active	ru	\N	5	{2c02c015-bfc5-479a-90ea-6d7c1a413670}	0	2026-08-18 09:03:03.60712+00	2026-08-18 09:03:19.441099+00	\N
7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	Samandar	Umarjanov	male	+998507112430	\N	passenger	none	active	ru	\N	5	{e3b27967-b9c9-4276-8ab0-2a52ca52e27b}	0	2026-08-18 14:48:46.958009+00	2026-08-18 14:49:01.306053+00	\N
60513e3b-5f2a-4b2c-a37b-6e726c089d00	Новый	Пользователь	\N	+998991234567	\N	passenger	pending	active	ru	\N	5	{00a762c8-fadf-4fe7-84a0-cd2814940d11,1f6bdae4-44a9-4573-9d4c-4c025b8284de,0c18569d-d486-4d5a-9616-149d90cfb515,1e4e7f40-efdd-4c80-85fb-c4a08fe04df9,c00b1ac8-803d-4079-9a48-e3cd7ce32fa5,ebf22b49-855b-4917-994f-adcc8b99c8a4,71ddae33-ba93-4a2f-9596-3593fb668ec0,c350ac44-23bf-44e0-bee6-94a146a7d8b9,cd2b1178-b9a9-4092-914d-39d633865abe,8f67dc8e-2638-4e1a-ac67-fe493742843f,ac5f31bc-0c5e-4ed7-976f-ce03dade7a21,d41aa92a-d406-4cd1-b5bd-bf6d10fdd4d5,b2cc0a47-713a-44af-ba26-1968dd29af45,22c34022-4747-4cf4-a4dc-5cfa84804831,eaf2e4da-a33f-4b52-9ad6-abac5cd8f8c0,0ff049b6-c857-41f0-82de-726a69310be9,8e2b605e-3a69-4b9e-9e89-a3cd85ac67c8,4bfb2cfe-e5c8-43ec-8520-9ea965633867}	0	2026-08-19 00:34:10.555241+00	2026-08-21 20:51:56.95222+00	\N
2c675db3-8449-4700-a80e-7299ae30fdef	Jahongir	Yoqubjanov	male	+998931632632	\N	passenger	none	active	ru	\N	5	{d0868329-208b-42b7-af4f-d02883c52d5b}	0	2026-08-18 03:26:43.296291+00	2026-08-18 03:26:56.503612+00	\N
64c51ff8-e296-4048-958c-9a6edde73141	Amirjon	Axmedov	male	+998949366333	\N	passenger	none	active	ru	\N	5	{1b41a48d-bec3-4760-aa3a-b6912dfdbc66}	0	2026-08-18 17:53:34.99025+00	2026-08-18 17:53:52.281431+00	\N
3ec33233-74d4-4bc4-945d-f17843ac3ad8	Muazzam	Xamrabayeva	female	+998902228230	\N	passenger	none	active	ru	\N	5	{9916b2f6-4afa-44a7-9c76-f6079975d017}	0	2026-08-18 03:37:37.323425+00	2026-08-18 03:37:55.366685+00	\N
fb458c0f-3fe8-4675-b8f4-f08b113a289e	Ilhom	Turabayev	male	+998991444989	\N	passenger	none	active	ru	\N	5	{2b7e7c22-64fc-4bdb-aa91-e9168b5f9fe4}	0	2026-08-20 04:56:05.638593+00	2026-08-20 04:56:16.015145+00	\N
021e1311-a08f-494b-a45d-8bf788c3f7ba	Islom	Maxmudov	male	+998909470106	\N	passenger	none	active	ru	\N	5	{f43bf8a4-0561-4590-8a2b-5fad0c495683}	0	2026-08-18 08:59:33.372099+00	2026-08-18 08:59:44.045497+00	\N
5e6e5cc0-4533-47e0-b2a7-73d79bc2a88e	Sanjar	Nuriddinov	male	+998949448606	\N	passenger	none	active	ru	\N	5	{49d6193d-8a92-4a5c-b5e5-b082816c4bbd}	0	2026-08-18 18:01:48.742597+00	2026-08-18 18:02:04.151672+00	\N
5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	Malika	Tillabayeva	female	+998938559313	\N	passenger	none	active	ru	\N	5	{38feaeb3-aae2-4bc6-a8b6-e4e8014c71c6}	0	2026-08-18 13:37:21.302143+00	2026-08-18 18:36:07.796258+00	\N
53cfa0e5-c3ab-486f-8ae7-7dece34a1274	Hamraboyev	Zafarjon	male	+998905528204	\N	passenger	none	active	ru	\N	5	{2b643ef9-64b8-4835-9541-53995191d62b}	0	2026-08-19 06:06:49.318727+00	2026-08-19 06:07:02.711742+00	\N
f5db4703-e0f8-4ec4-8b19-0575fb5a4e38	ОЙБЕКЖОН	Каримов	male	+998996437305	\N	passenger	none	active	ru	\N	5	{bcbb7c92-65a1-4716-9f38-02c9d5871b44}	0	2026-08-19 17:32:43.258805+00	2026-08-19 17:33:06.024241+00	\N
d4af5780-70b6-4bae-a067-3a460a489b9a	Yulduzxon	Karimova	female	+998974291077	\N	passenger	none	active	ru	\N	5	{8ab3f184-89fa-48da-a955-2bba66089e17}	0	2026-08-19 15:29:58.382886+00	2026-08-19 15:30:18.903145+00	\N
d6219547-9bcb-455b-bcaa-61f95c1fca64	Rohimjon	A&#39;lohanov	male	+998917589808	\N	passenger	none	active	ru	https://api.yolyolakay.xyz/yolyolakay-media/b2d1abba-5b5f-4a74-bc1d-b01a8d543169.jpg	5	{3d32c1a4-cc60-4d01-a22e-5e2d4993f4b2}	0	2026-08-18 01:28:44.29598+00	2026-08-20 03:17:00.905623+00	\N
8bc95322-9fc9-4903-b893-79c37aa115bc	Oybekбек	KARIMOV	male	+998938122399	$2a$10$IAKObawoMchiMqH3jhoXH.RIR6W00AvOIcwKvYw3un.VWblQuB7yO	passenger	none	active	ru	https://api.yolyolakay.xyz/yolyolakay-media/da38f6ca-3706-4a99-89bc-22f747daf56b.jpg	5	{e4758ee8-058e-4e11-b9a8-e664ffec4fb2,41480b7c-9094-4e66-a719-c19916affd43,bcbb7c92-65a1-4716-9f38-02c9d5871b44}	0	2026-08-18 01:09:30.646192+00	2026-08-20 04:40:42.241871+00	\N
67a1ef47-5ea2-4ed6-af06-5e82a13678dc	Abdurahmon	Alokhonov	male	+998976497722	$2a$10$HEGDOHXMMaQfKQyp4r1rH./k92DZ65SFOdkmLpBPKuUThCl7namVG	passenger	none	active	ru	https://api.yolyolakay.xyz/api/v1/media/b1b99923-9af2-4e91-856f-8687d44fad21.jpg	5	{65b00172-420e-4ba3-b4df-361ee0f3578c,8a71fc8e-780a-4e56-af75-aa6954273fa7}	0	2026-08-18 04:36:02.486138+00	2026-08-21 14:20:49.518617+00	\N
dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	Muxammad	Raxmatillayev	male	+998993912512	$2a$10$g6iCIsjzbh6uuv3Pxs93wesjaw3jsPDYLWYk.iPy5ppoq1IOcnTN2	driver	approved	active	en	\N	5	{598410d5-0fee-426b-aa9e-b1e6eddb19ee,fde26fb8-a956-456c-8c26-05f2d18fe0b6}	0	2026-08-18 01:07:14.716316+00	2026-08-21 13:08:04.188539+00	\N
9b30a80b-3940-4810-9b00-855ee44baeb2	Аброр	Керимов	male	+998910479089	\N	passenger	none	active	ru	\N	5	{804c8116-1012-49f9-8b09-17a427111c13}	0	2026-08-21 19:05:32.531822+00	2026-08-21 19:05:41.673665+00	\N
\.


--
-- Name: cars cars_gov_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cars
    ADD CONSTRAINT cars_gov_number_key UNIQUE (gov_number);


--
-- Name: cars cars_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cars
    ADD CONSTRAINT cars_pkey PRIMARY KEY (id);


--
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (id);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: user_blocks user_blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_blocks
    ADD CONSTRAINT user_blocks_pkey PRIMARY KEY (blocker_id, blocked_id);


--
-- Name: user_documents user_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_documents
    ADD CONSTRAINT user_documents_pkey PRIMARY KEY (id);


--
-- Name: users users_phone_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_cars_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cars_user_id ON public.cars USING btree (user_id);


--
-- Name: idx_reviews_users; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reviews_users ON public.reviews USING btree (reviewer_id, reviewee_id);


--
-- Name: idx_user_documents_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_documents_user_id ON public.user_documents USING btree (user_id);


--
-- Name: idx_users_phone; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_phone ON public.users USING btree (phone_number);


--
-- Name: cars cars_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cars
    ADD CONSTRAINT cars_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reviews reviews_reviewee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_reviewee_id_fkey FOREIGN KEY (reviewee_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reviews reviews_reviewer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_reviewer_id_fkey FOREIGN KEY (reviewer_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_documents user_documents_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_documents
    ADD CONSTRAINT user_documents_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

