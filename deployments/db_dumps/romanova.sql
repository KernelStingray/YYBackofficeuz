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

ALTER TABLE IF EXISTS ONLY public.parcel_responses DROP CONSTRAINT IF EXISTS parcel_responses_parcel_id_fkey;
DROP INDEX IF EXISTS public.idx_parcels_status;
DROP INDEX IF EXISTS public.idx_parcels_origin;
DROP INDEX IF EXISTS public.idx_parcels_destination;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.parcels DROP CONSTRAINT IF EXISTS parcels_pkey;
ALTER TABLE IF EXISTS ONLY public.parcel_responses DROP CONSTRAINT IF EXISTS parcel_responses_pkey;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.parcels;
DROP TABLE IF EXISTS public.parcel_responses;
DROP EXTENSION IF EXISTS postgis;
--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry and geography spatial types and functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: parcel_responses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.parcel_responses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    parcel_id uuid NOT NULL,
    driver_id uuid NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    offered_price numeric,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: parcels; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.parcels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    sender_id uuid NOT NULL,
    driver_id uuid,
    origin public.geometry(Point,4326) NOT NULL,
    destination public.geometry(Point,4326) NOT NULL,
    receiver_phone_encrypted character varying(255) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    price numeric NOT NULL,
    origin_address character varying(500),
    destination_address character varying(500),
    description text,
    size character varying(50),
    deliver_by timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
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
-- Data for Name: parcel_responses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.parcel_responses (id, parcel_id, driver_id, status, offered_price, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: parcels; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.parcels (id, sender_id, driver_id, origin, destination, receiver_phone_encrypted, status, price, origin_address, destination_address, description, size, deliver_by, created_at, updated_at, deleted_at) FROM stdin;
e4b28b74-19de-4453-8add-2fef2045b05d	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E61000006519E258175751400BB5A679C7814440	kVb59H6bjOOaVny79Kr0pAB9YpVOa/cmWRa5+Kdv2Q1w+ganP3Odvtk=	pending	25000.000000	Toshkent	Nurafshon	Деньги	small	2026-08-19 19:00:00+00	2026-08-18 09:24:52.639303+00	2026-08-18 09:24:52.639303+00	\N
d081a925-45c8-4f4b-abaf-301c8eb5bec4	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E61000006519E258175751400BB5A679C7814440	0101000020E610000043AA285E654F514018B49080D1A74440	lE4TruOGCVBJiJAO/pKwAWTbZmPOGENtcHhm5wJjab/cKDlSaqv++gk=	pending	25000.000000	Nurafshon	Toshkent	Кукат	medium	2026-08-19 19:00:00+00	2026-08-18 18:31:49.815956+00	2026-08-18 18:31:49.815956+00	\N
d44a6710-711d-492c-99dc-23d6032574c9	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	\N	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E6100000865AD3BCE3305140ABCFD556EC774440	Qrr8UvwB00YZG/ZXtdRyUbefW3OycBisrrviwe2kdxpq4P6Pk6uSty0=	pending	50000.000000	Toshkent	Chinoz	Goz	medium	\N	2026-08-18 18:35:42.326814+00	2026-08-18 18:35:42.326814+00	\N
3a62310d-16e7-4347-9e2f-2a28031c81a1	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E61000006519E258175751400BB5A679C7814440	NIzCSOiu7DgNn3X3Bl0zl9vUQ4hB6p4oXKzL8CFFBIL84Xhq7mE3Dmg=	pending	50000.000000	Toshkent	Nurafshon	5 кило узум	medium	\N	2026-08-19 01:53:03.157638+00	2026-08-19 01:53:03.157638+00	\N
c92a339a-1199-41d8-8017-354fbd85ae0d	d6219547-9bcb-455b-bcaa-61f95c1fca64	\N	0101000020E610000029ED0DBE30E951408D28ED0DBE804440	0101000020E610000043AA285E654F514018B49080D1A74440	A9m40NUsCmITRQMHnBB6fR8KgUFggznjOToLjIqtjrJsyG/7ephpUyU=	pending	1200.000000	Namangan	Toshkent	Mushuk	large	2026-08-20 19:00:00+00	2026-08-19 07:04:15.563995+00	2026-08-19 07:04:15.563995+00	\N
5a0eba51-93e0-4bf9-930b-881d72771ac0	5e6e5cc0-4533-47e0-b2a7-73d79bc2a88e	\N	0101000020E6100000188A856BFF565140FDEB264234844440	0101000020E6100000287E8CB96BBD504088855AD3BCD34340	3lwrEqrrRtXMzDQZUkcpaskG64UCY3LBDRLbXfdAJ+k3ZhBuNr9QwLg=	pending	5000.000000	Qumariq	Samarqand	Анор 5 кг	medium	2026-08-19 19:00:00+00	2026-08-19 11:57:11.10619+00	2026-08-19 11:57:11.10619+00	\N
8abb824d-7766-476e-8d63-51385b1b1388	64c51ff8-e296-4048-958c-9a6edde73141	\N	0101000020E6100000B956D691FE56514014E0055B36844440	0101000020E610000043AA285E654F514018B49080D1A74440	jAz0lwHRoSGm0RLmMfbeZkruqm+iPgy3tI5TZBQU1Q1AnLhBgXoiRGE=	pending	25000.000000	Qumariq	Toshkent	Cemodan	medium	\N	2026-08-19 12:38:50.422524+00	2026-08-19 12:38:50.422524+00	\N
801e7bf7-d536-4f05-89a1-e829d9f275de	021e1311-a08f-494b-a45d-8bf788c3f7ba	\N	0101000020E6100000F16F86D1E256514024592CEA38844440	0101000020E6100000F54A598638CE4D40C2172653053B4540	T77hP7/w5Jxw8PTUfYE0AOymjjMpDxnSv0OezEWjVh3NBYlKZ1ztS0w=	pending	100000.000000	Qumariq	Nukus	Zarif	medium	2026-08-26 19:00:00+00	2026-08-19 12:51:11.33053+00	2026-08-19 12:51:11.33053+00	\N
19bc2866-bb14-4e44-991b-09fb2a0180d1	d4af5780-70b6-4bae-a067-3a460a489b9a	\N	0101000020E610000037FE44654355514042DEBC827E994440	0101000020E6100000B81E85EB51BC51403EE8D9ACFA444440	TX1o4KKg5aKENlYYxjIInvVTTqCc49oUjSeKLFBiZnUns6H3Prs5QcE=	pending	50000.000000	Toshkent	Qo'qon	Dori	small	2026-08-21 19:00:00+00	2026-08-19 15:31:20.197129+00	2026-08-19 15:31:20.197129+00	\N
51fcf8bd-3a85-4142-9154-fa08d807e8fe	d4af5780-70b6-4bae-a067-3a460a489b9a	\N	0101000020E6100000287E8CB96BBD504088855AD3BCD34340	0101000020E61000006DE7FBA9F11A50405F07CE1951E24340	aIS7BfeAErEsWNLUv2KchCLOjPYBi4zKVs6kMjaS0ETyW9AyyPHaoyk=	pending	45000.000000	Samarqand	Buxoro	Xujjat	small	2026-08-19 19:00:00+00	2026-08-19 15:35:07.789691+00	2026-08-19 15:35:07.789691+00	\N
87642723-73df-4e64-adc6-e87ba8dee9ee	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E6100000A87B5347FD5651405FB35C363A844440	0101000020E61000004DE779266AF4514065A6594524894440	M5v00i5jqa2gHJt7bFzlgC6T8Oe+/KK2vmFg7UPFjGDvJtfU31VpANk=	canceled	55000.000000	Qumariq	Chortoq	134	small	\N	2026-08-19 16:58:24.587324+00	2026-08-20 01:47:51.299032+00	\N
7d7f58a3-7868-4507-8ef8-e77a2cf11ad2	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E61000000D897B2CFD565140595249F839844440	0101000020E610000085EB51B81E0D5040C3D32B6519CA4440	70+pPtXklEs7zdvt/FrEYjsH1P02FIjCD9WdLlOFIchXwuZxnf0j/mI=	canceled	55000.000000	Qumariq	Zarafshon	Картон	large	2026-08-21 19:00:00+00	2026-08-19 02:56:24.331143+00	2026-08-20 01:48:01.284713+00	\N
4970821a-ab8e-4d22-820b-670adef90e72	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	\N	0101000020E610000029ED0DBE30E951408D28ED0DBE804440	0101000020E610000043AA285E654F514018B49080D1A74440	wzbny1Qlz9FY3OKBoFiRW9GvXBOsdiXGpnI47DbLsN1yjEHej/YLMsg=	pending	100.000000	Namangan	Toshkent	Pul	small	\N	2026-08-20 04:18:23.502117+00	2026-08-20 04:18:23.502117+00	\N
8ef36769-32a4-408c-ac32-9700dc4f1fec	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E6100000F5B3A217FF565140A5EE14BC32844440	0101000020E6100000287E8CB96BBD504088855AD3BCD34340	gRtGp8U+svCbkP9nD6aFXs8wjidPdkN/n0YkIMd2c2tHPp65sFIpBVs=	pending	35000.000000	O'rtachirchiq	Samarqand	Доки	small	2026-08-21 19:00:00+00	2026-08-20 04:39:38.381946+00	2026-08-20 04:39:38.381946+00	\N
df8a4439-4f22-45b9-8129-da07ce45c696	fb458c0f-3fe8-4675-b8f4-f08b113a289e	\N	0101000020E61000006519E258175751400BB5A679C7814440	0101000020E6100000287E8CB96BBD504088855AD3BCD34340	pGPAymrryhnzO2Ipmi2r5cYR3oNGI7E+tr2Puz629T1Y/LkJTmu8BAY=	pending	25000.000000	Nurafshon	Samarqand	Hujjatlar	small	\N	2026-08-20 04:57:13.157712+00	2026-08-20 04:57:13.157712+00	\N
211507fc-a95c-4d3a-a656-0be9df09a340	8bc95322-9fc9-4903-b893-79c37aa115bc	\N	0101000020E6100000B40F689FFD565140CCE550D037844440	0101000020E6100000287E8CB96BBD504088855AD3BCD34340	miI45peNEKPkleKASol4yCopcwGgGDOxSQoRK+zyvzREna2i9fMKPC4=	canceled	25000.000000	O'rtachirchiq	Samarqand	Бук⁹	small	2026-08-20 19:00:00+00	2026-08-20 12:33:48.308378+00	2026-08-20 12:34:10.029128+00	\N
89ff1015-9254-48ed-a125-cc742de12884	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	\N	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E610000029ED0DBE30E951408D28ED0DBE804440	wFkcGWM7FJDWrATlVkaznFgPxRHdmBf+6JqvaZf+6tv3hg3NhLXi5dQ=	pending	100.000000	Toshkent	Namangan	Kiyim	medium	2026-08-29 19:00:00+00	2026-08-21 09:25:31.401626+00	2026-08-21 09:25:31.401626+00	\N
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
-- Name: parcel_responses parcel_responses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.parcel_responses
    ADD CONSTRAINT parcel_responses_pkey PRIMARY KEY (id);


--
-- Name: parcels parcels_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.parcels
    ADD CONSTRAINT parcels_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_parcels_destination; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_parcels_destination ON public.parcels USING gist (destination);


--
-- Name: idx_parcels_origin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_parcels_origin ON public.parcels USING gist (origin);


--
-- Name: idx_parcels_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_parcels_status ON public.parcels USING btree (status);


--
-- Name: parcel_responses parcel_responses_parcel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.parcel_responses
    ADD CONSTRAINT parcel_responses_parcel_id_fkey FOREIGN KEY (parcel_id) REFERENCES public.parcels(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

