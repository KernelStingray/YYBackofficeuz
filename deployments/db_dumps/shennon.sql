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

DROP INDEX IF EXISTS public.idx_events_event_type;
DROP INDEX IF EXISTS public.idx_events_created_at;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.events DROP CONSTRAINT IF EXISTS events_pkey;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.events;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id uuid NOT NULL,
    event_type character varying(100) NOT NULL,
    user_id uuid,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
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
-- Data for Name: events; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.events (id, event_type, user_id, payload, created_at, deleted_at) FROM stdin;
7cbb9b42-5fc3-402b-a6e6-044fc266d154	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:18.007988+00	\N
592f4676-b2d5-4998-9deb-0487141dff1f	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:19.663913+00	\N
e121900d-749e-40f9-9422-39b9945d8ca3	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:21.142705+00	\N
5f5c8fc6-648a-4e17-82e0-1798714dca6d	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:22.629776+00	\N
35e506fd-8659-4893-9ccf-8d3058f29a9f	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:27.964684+00	\N
01eaea04-4a5b-4b77-bf8f-057b52d44372	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:29.183677+00	\N
84d0ca02-0610-40c6-8a33-f9dcd6444d2a	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:30.767907+00	\N
9cfa80e0-8942-426f-bcbd-448ba603a671	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:32.301543+00	\N
a74d9aea-9d38-4d9c-a223-2d2f0e04f8a7	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:33.821652+00	\N
4ab7d196-dbd6-4547-92c1-b975284d98d0	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:37.003672+00	\N
be59de53-d8b3-4e83-bebf-664d2367ed1d	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:38.602078+00	\N
b14f70b4-dc2b-41ec-b7c1-3e706e0b9254	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:44.441751+00	\N
d3d86f3c-206c-4204-856e-108c3abfbd35	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:12:46.065904+00	\N
56d89923-4c54-457f-a2ef-adc161bac829	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:00.862464+00	\N
72ffafa0-cc49-44d3-9948-ebe9a468a3eb	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:13.559386+00	\N
db696b50-de24-40cd-8251-797bad6a9418	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:14.972823+00	\N
bfe97586-990d-4b81-a8d4-ec001df84ecd	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:41.340773+00	\N
36a132b4-a7ec-42d4-aa23-5c7fe4ef8603	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:43.236232+00	\N
43b39f5a-a032-4bc8-a532-71a2a72b3030	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:44.943148+00	\N
2594b734-5f70-4f28-b1ce-7a662e8eb9dc	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:47.333292+00	\N
5e3ea19e-4d88-493b-92c1-187691764333	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:49.034271+00	\N
7baec49b-7807-4e1f-a8a8-34382b50c704	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:51.164329+00	\N
9d056191-b0fb-4910-943f-94259c48a6f5	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:53.481624+00	\N
ef7cae31-8bca-4d34-91b7-8541ffbdde32	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:55.16901+00	\N
eb5b018b-2aed-44c0-a2af-0383afa364f9	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:57.012807+00	\N
bd7c4488-1baf-44b6-b424-f2f4fbf6db8c	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:13:59.440649+00	\N
0f426cf5-8ad5-4ab4-b674-78adf9c8aed8	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:01.144844+00	\N
6db9dfd6-dde6-4b3a-9a7b-2d9d29839833	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:02.731236+00	\N
4932a7fb-f9d9-4709-a89c-0bd32b132ac3	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:04.766211+00	\N
641e2f47-252a-4fc6-9b87-428e0c505c14	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:06.437559+00	\N
0c330a0a-c990-43cc-b96a-33c630345101	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:10.247458+00	\N
1b30bafa-ffd7-4cb8-99dc-8739b704ff91	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:12.250647+00	\N
e60932fb-2a2a-42bf-ad1f-f9c5ea3ea818	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:13.942873+00	\N
f3447e59-32b2-42af-8cc9-fe2dee9f12e7	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:15.651907+00	\N
c72f5d3a-1101-4c2e-be90-f824c0eee256	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:19.688876+00	\N
ffbd3e5c-77c0-4664-8d5f-3a22df4865e0	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:25.386016+00	\N
dfc72412-b127-467c-a9b2-dcec91e46e77	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:26.918583+00	\N
51ba6f79-0cd5-4da6-b233-3f18e306b8bc	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:29.46817+00	\N
0e2c9683-55b1-45f6-85ef-6b42553b395d	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:30.774932+00	\N
d8ce28ed-d4d7-4f19-a75c-2922f0ca4bff	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:32.251053+00	\N
49ed51c8-8e2a-4964-8ba2-28e901d059ca	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:33.963823+00	\N
e84c1403-3d72-4c83-93b3-305abcadd535	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:14:35.527516+00	\N
181aa7c1-f51a-4cf4-a1ca-6e53a5e06c67	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 01:16:49.015568+00	\N
c97cf283-859a-491c-bcbc-a38d0447729c	ride_published	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	{"seats": 3, "activated": true}	2026-08-18 01:26:00.210512+00	\N
5dbbc049-b87a-4c75-ab1b-6994e414d817	ride_published	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	{"seats": 3, "activated": false}	2026-08-18 01:26:24.356709+00	\N
67b92057-8036-4b12-bb25-819bf7162998	search_performed	3ec33233-74d4-4bc4-945d-f17843ac3ad8	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 03:38:55.112306+00	\N
3edf34fe-aeb9-47d1-95d7-7776fafac5cd	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 08:39:43.877537+00	\N
305f94b8-708b-4294-88fa-48deb0b65e86	search_performed	021e1311-a08f-494b-a45d-8bf788c3f7ba	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 08:59:59.918006+00	\N
3a601dc4-6fb5-47f5-ad01-24de8cc582e4	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 09:22:50.893655+00	\N
20282d66-cdc2-4f43-a95e-8ed23dd81856	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 09:24:03.08628+00	\N
59037726-9810-43c9-9f18-3acdf10fa998	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 2, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 13:38:28.723151+00	\N
b72eb4d6-f5c4-42b3-94cd-561d0e4249f4	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 14:49:44.1577+00	\N
7d3acaae-a232-40a0-b993-f79f13e0dcc6	search_performed	3ec33233-74d4-4bc4-945d-f17843ac3ad8	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-18 15:37:58.846716+00	\N
23ed9e55-b4b3-4fc9-b208-9c69c091d279	search_performed	3ec33233-74d4-4bc4-945d-f17843ac3ad8	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 15:38:20.321257+00	\N
a0db3fcb-1661-4eb7-8964-696400cdb394	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 18:34:48.567395+00	\N
17c6ea1d-b5f6-4a20-b289-a124a4f4997d	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-18 18:38:34.421766+00	\N
f48bb66b-5bde-4c3a-a946-6e49390f081f	plan_purchase_started	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"plan": "plus", "price": 5000}	2026-08-18 20:21:21.487558+00	\N
29bdf147-6490-427e-85b9-f3bd3dca83e8	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 00:02:47.807346+00	\N
b6fcea6e-b3f6-4aed-b531-b5d218e62cbe	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 00:02:52.392971+00	\N
73f93c8a-8142-47e2-85f9-31ce28a5988c	ride_opened	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"negotiable": true}	2026-08-19 00:03:15.333013+00	\N
84d52e3c-9cc2-45b5-b755-01374fe84d21	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 00:03:27.081122+00	\N
9692bd5b-56b9-4427-bdbd-b862fe9cbb47	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 00:03:29.986859+00	\N
c6561439-c67c-469d-ac83-52027d99ca07	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 00:28:41.114637+00	\N
a6b9208e-d97a-4eec-8f24-3df4ab86a13a	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 00:28:42.496068+00	\N
ec949210-9a4c-4ddb-8c88-d1080ddd8dc1	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 00:28:45.899593+00	\N
90d86f18-b29c-40be-bc28-a41f5f70f6e2	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 00:28:50.869083+00	\N
53e4eff2-4fe5-4f4a-80a2-93869d22f998	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 01:51:35.897514+00	\N
30c87be8-a63b-43e6-ae8c-2549ae5b9ed2	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:03:08.069077+00	\N
405a3753-3610-469a-828c-646b964cb539	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 02:03:10.381195+00	\N
fd9ca9a3-bf83-47a3-a110-b47f56af67a1	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:03:12.334572+00	\N
480f55fc-6a59-4067-b705-b09891793165	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:03:14.026891+00	\N
6500bd70-68dc-4d42-bc0d-6a1e0f60bfa5	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 02:03:15.446494+00	\N
b77e206d-9a6d-4880-b69d-dea752f4d478	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 1, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:03:16.226757+00	\N
ce24940a-a7fe-4260-9365-7bea548081a5	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:03:21.942392+00	\N
5f06dbf7-51e1-48bf-b52a-8b8af6c45f01	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:05:32.796199+00	\N
03c57b24-ce14-4cdb-a5da-e333960acc13	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 02:05:34.547372+00	\N
5f331c4b-6d6d-4985-a005-d100f968e0fb	ride_opened	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"negotiable": true}	2026-08-19 02:05:36.7012+00	\N
56f3f1ac-0884-466d-9bd3-5e98f407005c	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 1, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:05:49.645166+00	\N
c7aaa50b-bd9a-4313-8425-11adc3eeeeb6	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:05:59.671835+00	\N
5d793e95-0ad1-470a-8e24-9d00496b89b1	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 1, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:06:51.38002+00	\N
8313b2be-9ad2-4920-9c38-049b7b0d9c72	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:06:52.548688+00	\N
0b7b1710-6ff1-45b7-adbe-81ff5eb3b2cf	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 1, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:07:22.341007+00	\N
45303fa8-a83a-4cf3-8f03-0247fec966f1	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:07:52.851948+00	\N
b75adf66-605f-4ad9-8042-2180d1526d77	booking_created	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 3, "buyout": true, "negotiated": true}	2026-08-19 02:08:48.649472+00	\N
d7270a33-18e9-42c9-ab68-1c9389ae2366	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:09:52.57977+00	\N
ad5318a7-a753-47f2-b352-1e6b5bca7245	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 1, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 02:09:54.167045+00	\N
4f082db8-2195-4396-b912-5c722cb55c15	ride_opened	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"negotiable": true}	2026-08-19 02:09:55.111706+00	\N
4791212e-6265-4fb7-a95c-3ce829bb954d	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 3, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:12:25.325641+00	\N
2c61256d-3fc1-4a9c-9787-92078b816edd	search_performed	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"seats": 3, "failed": false, "results": 1, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 02:12:26.81018+00	\N
0e253701-611a-4863-901d-5c6e2e8a9222	ride_opened	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"negotiable": true}	2026-08-19 02:12:27.934006+00	\N
df13fc76-aa24-49a5-b343-a8fa2d5a0e2a	ride_opened	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"negotiable": true}	2026-08-19 02:12:54.333362+00	\N
d12457b4-a873-4740-bc27-42224e92a945	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:18:20.615071+00	\N
877f7037-983a-4fb5-b781-b8658e5c7bae	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:18:31.208708+00	\N
e939dafa-4f68-4101-9f22-28fc6f08b44d	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:18:42.026425+00	\N
302ba346-dfa8-4490-9d03-8a499c22d17b	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:40:00.903232+00	\N
836f1592-ea45-40dc-908b-4de0530d51ce	booking_created	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 3, "buyout": true, "negotiated": false}	2026-08-19 02:40:17.178639+00	\N
b1397c38-29f9-4440-871c-c2d3354cab62	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:43:53.753214+00	\N
42214964-c258-4d3b-8731-58fd0f8f6d61	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 02:54:36.536193+00	\N
fbffaec9-3b12-4a2a-a10d-cd6c85327be2	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 02:54:46.553254+00	\N
1e5d7108-5c86-4895-85e5-94fc8dd8456b	search_performed	53cfa0e5-c3ab-486f-8ae7-7dece34a1274	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 06:07:42.812324+00	\N
68df7e6b-e425-40a4-9b9f-ccec204c2b65	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 10:46:00.46278+00	\N
9d8c921e-a009-46f3-a3bb-f25c3f2d91db	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 10:46:07.497363+00	\N
bf62bb38-0a1c-490c-b9c6-e93656414dd3	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 10:46:12.647659+00	\N
64de2c50-c25d-4714-9e5b-7a327d218376	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 10:46:20.889272+00	\N
81e80419-98a2-467a-8603-953784ad73f1	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 10:46:48.468851+00	\N
fef9f3d3-3395-47f3-af78-985f3bfdcb63	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 10:46:58.899765+00	\N
d1c45965-7fdc-4bb3-a6eb-ecfa902024ee	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 10:46:59.550526+00	\N
26f487ea-6eb2-47b2-b381-40a6f702bdaf	search_performed	5e6e5cc0-4533-47e0-b2a7-73d79bc2a88e	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 11:54:14.980729+00	\N
4ec51d5d-f429-4e68-9e5d-d3d6ff8d68df	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 12:43:50.294359+00	\N
5c60667d-bddd-4337-b037-bb48f8018c0c	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 12:45:08.34949+00	\N
aa18ff4a-66d1-42d6-b78d-d7ec460d94e7	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 12:45:09.889195+00	\N
548c5495-7f13-4eee-a2d8-a8f8105b4d44	search_performed	d4af5780-70b6-4bae-a067-3a460a489b9a	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 15:31:50.817395+00	\N
753a70e5-341b-4fb4-969e-c9b785c5cb1f	search_performed	d4af5780-70b6-4bae-a067-3a460a489b9a	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 15:31:56.399323+00	\N
2b90ba03-0cd1-4e2e-af9a-d75b7fada33e	search_performed	d4af5780-70b6-4bae-a067-3a460a489b9a	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 15:34:09.106652+00	\N
d6f39ae9-2d43-4d5f-9611-a297e96d69c2	search_performed	d4af5780-70b6-4bae-a067-3a460a489b9a	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 15:34:14.48535+00	\N
33e6288f-b099-41a7-acb3-20a4aac11dd0	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 16:56:34.237354+00	\N
992bf592-a810-48c2-8951-9043152a8257	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 16:56:42.608188+00	\N
45840c0d-d7ec-42ad-ba80-c850f276d1b7	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 16:56:58.110347+00	\N
d0c55b13-76d4-4c09-b1a4-4063f17d566d	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 17:29:02.908315+00	\N
c09ca615-1700-4770-9fac-112b11653e70	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 17:29:40.318455+00	\N
58f3cd80-7a3d-4483-81af-00e9ebfa5c31	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 2, "failed": false, "results": 0, "has_date": false, "radius_km": 100, "has_price_filter": false}	2026-08-19 17:29:46.442719+00	\N
d4a5145b-13c0-4ea3-b521-ba7be4904b4c	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-19 17:49:31.052478+00	\N
51be861f-58e7-4b7c-b380-ba1e54660c0e	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-19 18:28:39.746657+00	\N
d9a7c0ba-2cb8-4853-9683-58d71e0edb9f	ride_published	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	{"activated": false}	2026-08-20 00:57:08.462757+00	\N
6478b0e8-92b7-44d5-b6b7-a716a396a22c	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-20 01:48:20.644237+00	\N
6c10e90a-4fab-46ab-b477-e99419cb4d77	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-20 01:48:25.378688+00	\N
b891a090-e051-4f13-89db-61f7fbf4bf8b	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-20 01:48:27.250582+00	\N
9e13c2db-e73b-49f3-b907-942c7904d116	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-20 02:14:16.98081+00	\N
91a62f79-7fe1-4e98-a4bd-5a5edbf26b6d	search_performed	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-20 02:14:19.902271+00	\N
48af0fe3-3f87-4b03-8061-3763b9150511	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 2, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-20 04:17:19.043071+00	\N
c2d8ff95-3be2-4e0f-abe1-803c6bf94dcd	plan_purchase_started	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"plan": "premium", "price": 10000}	2026-08-20 04:19:14.201607+00	\N
6772c36d-c4bc-4a38-8db1-c13394ec2282	plan_purchase_started	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"plan": "plus", "price": 5000}	2026-08-20 04:20:06.603329+00	\N
37806fcf-85a0-483e-b020-017b1d91d543	plan_purchase_started	8bc95322-9fc9-4903-b893-79c37aa115bc	{"plan": "plus", "price": 5000}	2026-08-20 04:41:03.691278+00	\N
5f9dbb84-737b-4ca1-be3d-15b0606f4e46	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-20 04:42:16.470293+00	\N
394b2edb-7c6f-40c5-8211-720c09a68f35	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-20 04:44:43.898511+00	\N
b4e737f3-18ee-4dd7-bd3c-88efa92bdf70	search_performed	44b88e16-3feb-4bce-850f-0c2d4b47ff80	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-20 10:57:59.557543+00	\N
74f14ffb-0cf9-4aad-933e-319f1c8fc8c7	search_performed	44b88e16-3feb-4bce-850f-0c2d4b47ff80	{"seats": 1, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-20 10:58:10.586588+00	\N
c3a5bd90-3cc4-4245-8e93-5554751b1b34	plan_purchase_started	8bc95322-9fc9-4903-b893-79c37aa115bc	{"plan": "plus", "price": 5000}	2026-08-20 12:35:06.192752+00	\N
4b16386b-b7fc-4b75-97f0-30ff0f02af1c	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 2, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-20 17:08:35.814655+00	\N
19889529-ecfd-47ff-8eaa-5c82c34180d9	search_performed	3ec33233-74d4-4bc4-945d-f17843ac3ad8	{"seats": 1, "failed": false, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-21 08:51:30.314386+00	\N
59669267-2a37-42c5-aaea-ac795a7b8b33	search_performed	3ec33233-74d4-4bc4-945d-f17843ac3ad8	{"seats": 1, "failed": true, "results": 0, "has_date": false, "radius_km": 25, "has_price_filter": false}	2026-08-21 08:51:39.197615+00	\N
a57c6229-921b-49c3-a48b-79bd9c391fb8	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 2, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-21 09:24:39.876815+00	\N
5d4596f1-eec6-46d7-b922-bceb11a5a67d	search_performed	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	{"seats": 2, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-21 09:24:50.317216+00	\N
5f1d9dd0-8424-433a-8e6a-a8799bd0ac18	search_performed	8bc95322-9fc9-4903-b893-79c37aa115bc	{"seats": 3, "failed": false, "results": 0, "has_date": true, "radius_km": 25, "has_price_filter": false}	2026-08-21 14:28:29.297455+00	\N
8f434a97-b9bf-4f13-b36d-73f81434c357	ride_opened	8bc95322-9fc9-4903-b893-79c37aa115bc	{"negotiable": true}	2026-08-21 14:28:45.608011+00	\N
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_events_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_created_at ON public.events USING btree (created_at);


--
-- Name: idx_events_event_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_event_type ON public.events USING btree (event_type);


--
-- PostgreSQL database dump complete
--

