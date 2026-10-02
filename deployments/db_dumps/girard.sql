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

ALTER TABLE IF EXISTS ONLY public.user_promo_usages DROP CONSTRAINT IF EXISTS user_promo_usages_promo_code_id_fkey;
DROP INDEX IF EXISTS public.idx_user_subscriptions_user_id;
DROP INDEX IF EXISTS public.idx_user_promo_usages_user_id;
DROP INDEX IF EXISTS public.idx_transactions_user_id;
DROP INDEX IF EXISTS public.idx_transactions_external_tx_id;
DROP INDEX IF EXISTS public.idx_promo_codes_code;
DROP INDEX IF EXISTS public.idx_payment_tokens_user_id;
ALTER TABLE IF EXISTS ONLY public.user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_pkey;
ALTER TABLE IF EXISTS ONLY public.user_promo_usages DROP CONSTRAINT IF EXISTS user_promo_usages_user_id_promo_code_id_key;
ALTER TABLE IF EXISTS ONLY public.user_promo_usages DROP CONSTRAINT IF EXISTS user_promo_usages_pkey;
ALTER TABLE IF EXISTS ONLY public.transactions DROP CONSTRAINT IF EXISTS transactions_pkey;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.promo_codes DROP CONSTRAINT IF EXISTS promo_codes_pkey;
ALTER TABLE IF EXISTS ONLY public.promo_codes DROP CONSTRAINT IF EXISTS promo_codes_code_key;
ALTER TABLE IF EXISTS ONLY public.payment_tokens DROP CONSTRAINT IF EXISTS payment_tokens_pkey;
DROP TABLE IF EXISTS public.user_subscriptions;
DROP TABLE IF EXISTS public.user_promo_usages;
DROP TABLE IF EXISTS public.transactions;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.promo_codes;
DROP TABLE IF EXISTS public.payment_tokens;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: payment_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    provider character varying(50) NOT NULL,
    token character varying(255) NOT NULL,
    masked_pan character varying(20) NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: promo_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promo_codes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(50) NOT NULL,
    discount_amount numeric,
    discount_percent numeric,
    discount_type character varying(20) DEFAULT 'fixed_amount'::character varying,
    max_uses integer DEFAULT 1 NOT NULL,
    current_uses integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    starts_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone,
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
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    amount numeric NOT NULL,
    currency character varying(10) DEFAULT 'UZS'::character varying NOT NULL,
    type character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    provider character varying(50) NOT NULL,
    payment_token_id uuid,
    external_tx_id character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: user_promo_usages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_promo_usages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    promo_code_id uuid NOT NULL,
    used_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: user_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    plan_name character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying NOT NULL,
    start_date timestamp with time zone DEFAULT now() NOT NULL,
    end_date timestamp with time zone NOT NULL,
    has_used_trial boolean DEFAULT false NOT NULL,
    rides_left integer DEFAULT 0 NOT NULL,
    free_parcels_quota integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Data for Name: payment_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.payment_tokens (id, user_id, provider, token, masked_pan, status, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: promo_codes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.promo_codes (id, code, discount_amount, discount_percent, discount_type, max_uses, current_uses, is_active, starts_at, expires_at, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.transactions (id, user_id, amount, currency, type, status, provider, payment_token_id, external_tx_id, created_at, updated_at, deleted_at) FROM stdin;
07fe0d96-8020-4be9-96b9-8c91827b89ea	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	5000.000000	UZS	subscription:plus	cancelled	click	\N	\N	2026-08-18 20:21:21.517504+00	2026-08-18 20:52:00.035175+00	\N
45477a8f-aa39-4ad6-aa5a-78ec3a45cb5a	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	5000.000000	UZS	subscription	success	system_mock	\N	\N	2026-08-20 00:50:00.071518+00	2026-08-20 00:50:00.075767+00	\N
9cb7c4bf-546b-4328-ae9a-82e419c688b3	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	10000.000000	UZS	subscription:premium	cancelled	click	\N	\N	2026-08-20 04:19:14.392748+00	2026-08-20 04:50:00.480961+00	\N
a65991f7-0089-45fa-a9ac-4c680619034e	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	5000.000000	UZS	subscription:plus	cancelled	click	\N	\N	2026-08-20 04:20:06.804738+00	2026-08-20 04:51:00.861229+00	\N
ca1f7823-b79c-48f6-8094-35787fd7fd7e	8bc95322-9fc9-4903-b893-79c37aa115bc	5000.000000	UZS	subscription:plus	cancelled	click	\N	\N	2026-08-20 04:41:03.874867+00	2026-08-20 05:12:00.390389+00	\N
734b57f6-7bcb-46bb-be41-cf486a90e52f	8bc95322-9fc9-4903-b893-79c37aa115bc	5000.000000	UZS	subscription:plus	cancelled	click	\N	\N	2026-08-20 12:35:06.204126+00	2026-08-20 13:06:00.064353+00	\N
\.


--
-- Data for Name: user_promo_usages; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_promo_usages (id, user_id, promo_code_id, used_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: user_subscriptions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_subscriptions (id, user_id, plan_name, status, start_date, end_date, has_used_trial, rides_left, free_parcels_quota, created_at, updated_at, deleted_at) FROM stdin;
e4719303-dd13-4bb5-9e69-c6bf39a73fb8	8bc95322-9fc9-4903-b893-79c37aa115bc	trial_ride	active	2026-08-18 01:09:30.668851+00	2026-09-18 01:09:30.668851+00	t	1	0	2026-08-18 01:09:30.668996+00	2026-08-18 01:09:30.668996+00	\N
ff231db1-e8eb-46f9-aed4-93b713acee21	d6219547-9bcb-455b-bcaa-61f95c1fca64	trial_ride	active	2026-08-18 01:28:44.300583+00	2026-09-18 01:28:44.300583+00	t	1	0	2026-08-18 01:28:44.300605+00	2026-08-18 01:28:44.300605+00	\N
c462229f-3a36-4699-8d42-b7813dbd1e47	2c675db3-8449-4700-a80e-7299ae30fdef	trial_ride	active	2026-08-18 03:26:43.305891+00	2026-09-18 03:26:43.305891+00	t	1	0	2026-08-18 03:26:43.306043+00	2026-08-18 03:26:43.306043+00	\N
6eedbcb7-f2ba-4ea4-a7d9-dc579f3824f6	3ec33233-74d4-4bc4-945d-f17843ac3ad8	trial_ride	active	2026-08-18 03:37:37.328135+00	2026-09-18 03:37:37.328135+00	t	1	0	2026-08-18 03:37:37.328156+00	2026-08-18 03:37:37.328156+00	\N
e42a2899-f594-45a6-a74b-33bf901c3d0d	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	trial_ride	active	2026-08-18 04:36:02.492225+00	2026-09-18 04:36:02.492225+00	t	1	0	2026-08-18 04:36:02.492371+00	2026-08-18 04:36:02.492371+00	\N
e8d7d719-2531-4ad2-8d97-00dcffbf7d82	021e1311-a08f-494b-a45d-8bf788c3f7ba	trial_ride	active	2026-08-18 08:59:33.378043+00	2026-09-18 08:59:33.378043+00	t	1	0	2026-08-18 08:59:33.378184+00	2026-08-18 08:59:33.378184+00	\N
a22d5151-9fef-4f2d-848d-8040f9a6f342	da5a196d-c70d-4e0b-b0ea-e4133730f0cf	trial_ride	active	2026-08-18 09:03:03.612878+00	2026-09-18 09:03:03.612878+00	t	1	0	2026-08-18 09:03:03.613105+00	2026-08-18 09:03:03.613105+00	\N
43239f93-8eb0-47e4-81c1-018dcb57052b	5b8456a6-9539-4fe5-9c79-0a89b99a3b9b	trial_ride	active	2026-08-18 13:37:21.308192+00	2026-09-18 13:37:21.308192+00	t	1	0	2026-08-18 13:37:21.308398+00	2026-08-18 13:37:21.308398+00	\N
be718aac-ab2d-4a87-82a7-7ba5e73b20ca	7dfa81fa-ad4f-4f59-bb12-effdcd661ee3	trial_ride	active	2026-08-18 14:48:46.963591+00	2026-09-18 14:48:46.963591+00	t	1	0	2026-08-18 14:48:46.963733+00	2026-08-18 14:48:46.963733+00	\N
55bb3afa-8fb9-47a2-97bb-6e0eeb4ec225	64c51ff8-e296-4048-958c-9a6edde73141	trial_ride	active	2026-08-18 17:53:34.99519+00	2026-09-18 17:53:34.99519+00	t	1	0	2026-08-18 17:53:34.995324+00	2026-08-18 17:53:34.995324+00	\N
da26341c-a720-4eaa-a299-b1b3701f1757	5e6e5cc0-4533-47e0-b2a7-73d79bc2a88e	trial_ride	active	2026-08-18 18:01:48.756048+00	2026-09-18 18:01:48.756048+00	t	1	0	2026-08-18 18:01:48.756207+00	2026-08-18 18:01:48.756207+00	\N
09cdc65b-3854-467b-bdd7-0376fe99cb14	53cfa0e5-c3ab-486f-8ae7-7dece34a1274	trial_ride	active	2026-08-19 06:06:49.325964+00	2026-09-19 06:06:49.325964+00	t	1	0	2026-08-19 06:06:49.326091+00	2026-08-19 06:06:49.326091+00	\N
a45d79e8-8e88-4068-9437-e8c6b4a1cbe0	d4af5780-70b6-4bae-a067-3a460a489b9a	trial_ride	active	2026-08-19 15:29:58.388831+00	2026-09-19 15:29:58.388831+00	t	1	0	2026-08-19 15:29:58.388983+00	2026-08-19 15:29:58.388983+00	\N
6e61b30b-21fd-4d1a-8bce-45340a1f1d66	44b88e16-3feb-4bce-850f-0c2d4b47ff80	trial_ride	active	2026-08-19 15:35:14.58068+00	2026-09-19 15:35:14.58068+00	t	1	0	2026-08-19 15:35:14.580716+00	2026-08-19 15:35:14.580716+00	\N
70c307a7-c3c8-4d77-bfc9-a99cb7a8ae5a	f5db4703-e0f8-4ec4-8b19-0575fb5a4e38	trial_ride	active	2026-08-19 17:32:43.263891+00	2026-09-19 17:32:43.263891+00	t	1	0	2026-08-19 17:32:43.264049+00	2026-08-19 17:32:43.264049+00	\N
f1229b52-e03d-44ad-ad3d-f3447191d243	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	none	active	2026-08-18 01:07:15.017166+00	2026-09-20 00:49:36.382423+00	t	0	0	2026-08-18 01:07:15.017384+00	2026-08-20 00:50:00.07917+00	\N
6cf3b1c6-91cd-4e12-b683-14fc7ff3c604	fb458c0f-3fe8-4675-b8f4-f08b113a289e	trial_ride	active	2026-08-20 04:56:05.646662+00	2026-09-20 04:56:05.646662+00	t	1	0	2026-08-20 04:56:05.646823+00	2026-08-20 04:56:05.646823+00	\N
11936b03-9ea6-424c-88d3-4e73741a1b4e	61b756ea-dc74-4b96-afff-dc8e45f12424	trial_ride	active	2026-08-20 18:52:15.178431+00	2026-09-20 18:52:15.178431+00	t	1	0	2026-08-20 18:52:15.178768+00	2026-08-20 18:52:15.178768+00	\N
0b17eabc-980a-42ad-afaf-648b6339a01a	7f7aeb4d-2521-4a1f-96f0-cb5841bcd9db	trial_ride	active	2026-08-20 18:52:15.278255+00	2026-09-20 18:52:15.278255+00	t	1	0	2026-08-20 18:52:15.278278+00	2026-08-20 18:52:15.278278+00	\N
f0b70b34-937c-4744-9466-b794306500db	9732d852-9d54-4de7-a075-22ce9ba0e735	trial_ride	active	2026-08-20 18:52:15.374755+00	2026-09-20 18:52:15.374755+00	t	1	0	2026-08-20 18:52:15.374775+00	2026-08-20 18:52:15.374775+00	\N
df9bd29a-f38d-49b4-a7d5-a5e443681c4a	a0fd44ec-ce28-4adc-89b8-f103cb64f9bd	trial_ride	active	2026-08-20 18:52:15.475978+00	2026-09-20 18:52:15.475978+00	t	1	0	2026-08-20 18:52:15.476033+00	2026-08-20 18:52:15.476033+00	\N
dddd4ae2-a866-4b03-a76e-0884d2434512	326c8e40-e6c9-45f1-8a1d-033842d66d44	trial_ride	active	2026-08-20 18:52:15.577963+00	2026-09-20 18:52:15.577963+00	t	1	0	2026-08-20 18:52:15.578096+00	2026-08-20 18:52:15.578096+00	\N
8275119b-01a3-4157-8edb-46b2e97c160f	908cd936-c5b2-4c78-8d8f-ff16eb8f6715	trial_ride	active	2026-08-20 18:52:15.680427+00	2026-09-20 18:52:15.680427+00	t	1	0	2026-08-20 18:52:15.680487+00	2026-08-20 18:52:15.680487+00	\N
6c050846-2502-4e55-a78a-daa660a05726	48bd4254-2a45-407b-b4e4-cc2c57d8dfe0	trial_ride	active	2026-08-20 18:52:15.779243+00	2026-09-20 18:52:15.779243+00	t	1	0	2026-08-20 18:52:15.779288+00	2026-08-20 18:52:15.779288+00	\N
8627d4b3-0c1d-48c4-ae24-90ad29a09290	6784ec53-2abd-4e25-ab28-9ee3db39817a	trial_ride	active	2026-08-20 18:52:15.876209+00	2026-09-20 18:52:15.876209+00	t	1	0	2026-08-20 18:52:15.876235+00	2026-08-20 18:52:15.876235+00	\N
6f00ac94-9c58-4b11-b487-915a76baf834	82cf0bae-26c4-4094-8d70-e3f3a50048d2	trial_ride	active	2026-08-20 18:52:15.975587+00	2026-09-20 18:52:15.975587+00	t	1	0	2026-08-20 18:52:15.975613+00	2026-08-20 18:52:15.975613+00	\N
c7cdf3ff-c953-4665-b09e-934695fc9249	7580bf08-72d2-443e-8819-0ceacb2f83b1	trial_ride	active	2026-08-20 18:52:16.077784+00	2026-09-20 18:52:16.077784+00	t	1	0	2026-08-20 18:52:16.077814+00	2026-08-20 18:52:16.077814+00	\N
e48ca1ab-d036-4526-839b-d97ca4ee16de	9b30a80b-3940-4810-9b00-855ee44baeb2	trial_ride	active	2026-08-21 19:05:32.540398+00	2026-09-21 19:05:32.540398+00	t	1	0	2026-08-21 19:05:32.540543+00	2026-08-21 19:05:32.540543+00	\N
76921e5e-57b8-43c0-9f35-2c7d70a8d629	60513e3b-5f2a-4b2c-a37b-6e726c089d00	trial_ride	active	2026-08-19 00:34:10.596959+00	2026-09-19 00:34:10.596959+00	t	3	0	2026-08-19 00:34:10.597486+00	2026-08-21 20:53:51.225204+00	\N
\.


--
-- Name: payment_tokens payment_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_tokens
    ADD CONSTRAINT payment_tokens_pkey PRIMARY KEY (id);


--
-- Name: promo_codes promo_codes_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_code_key UNIQUE (code);


--
-- Name: promo_codes promo_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: user_promo_usages user_promo_usages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_promo_usages
    ADD CONSTRAINT user_promo_usages_pkey PRIMARY KEY (id);


--
-- Name: user_promo_usages user_promo_usages_user_id_promo_code_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_promo_usages
    ADD CONSTRAINT user_promo_usages_user_id_promo_code_id_key UNIQUE (user_id, promo_code_id);


--
-- Name: user_subscriptions user_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: idx_payment_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_tokens_user_id ON public.payment_tokens USING btree (user_id);


--
-- Name: idx_promo_codes_code; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promo_codes_code ON public.promo_codes USING btree (code);


--
-- Name: idx_transactions_external_tx_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_external_tx_id ON public.transactions USING btree (external_tx_id);


--
-- Name: idx_transactions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_transactions_user_id ON public.transactions USING btree (user_id);


--
-- Name: idx_user_promo_usages_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_promo_usages_user_id ON public.user_promo_usages USING btree (user_id);


--
-- Name: idx_user_subscriptions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_subscriptions_user_id ON public.user_subscriptions USING btree (user_id);


--
-- Name: user_promo_usages user_promo_usages_promo_code_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_promo_usages
    ADD CONSTRAINT user_promo_usages_promo_code_id_fkey FOREIGN KEY (promo_code_id) REFERENCES public.promo_codes(id);


--
-- PostgreSQL database dump complete
--

