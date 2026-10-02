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

ALTER TABLE IF EXISTS ONLY public.ride_stops DROP CONSTRAINT IF EXISTS ride_stops_ride_id_fkey;
ALTER TABLE IF EXISTS ONLY public.bookings DROP CONSTRAINT IF EXISTS bookings_ride_id_fkey;
DROP INDEX IF EXISTS public.idx_route_subs_passenger;
DROP INDEX IF EXISTS public.idx_route_subs_active;
DROP INDEX IF EXISTS public.idx_rides_status;
DROP INDEX IF EXISTS public.idx_rides_route_polygon;
DROP INDEX IF EXISTS public.idx_rides_origin;
DROP INDEX IF EXISTS public.idx_rides_destination;
DROP INDEX IF EXISTS public.idx_ride_stops_location;
DROP INDEX IF EXISTS public.idx_bookings_ride_id;
DROP INDEX IF EXISTS public.idx_bookings_passenger_id;
ALTER TABLE IF EXISTS ONLY public.sent_reminders DROP CONSTRAINT IF EXISTS unique_ride_user_kind;
ALTER TABLE IF EXISTS ONLY public.sent_reminders DROP CONSTRAINT IF EXISTS sent_reminders_pkey;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.route_subscriptions DROP CONSTRAINT IF EXISTS route_subscriptions_pkey;
ALTER TABLE IF EXISTS ONLY public.rides DROP CONSTRAINT IF EXISTS rides_pkey;
ALTER TABLE IF EXISTS ONLY public.ride_stops DROP CONSTRAINT IF EXISTS ride_stops_pkey;
ALTER TABLE IF EXISTS ONLY public.bookings DROP CONSTRAINT IF EXISTS bookings_pkey;
DROP TABLE IF EXISTS public.sent_reminders;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.route_subscriptions;
DROP TABLE IF EXISTS public.rides;
DROP TABLE IF EXISTS public.ride_stops;
DROP TABLE IF EXISTS public.bookings;
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
-- Name: bookings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bookings (
    id uuid NOT NULL,
    ride_id uuid NOT NULL,
    passenger_id uuid NOT NULL,
    seats_booked integer NOT NULL,
    total_price numeric(10,2) NOT NULL,
    is_full_car_buyout boolean DEFAULT false NOT NULL,
    offered_price numeric(10,2),
    status character varying DEFAULT 'pending'::character varying NOT NULL,
    pickup_preference character varying(30) DEFAULT 'wait_at_route'::character varying,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    CONSTRAINT bookings_seats_booked_check CHECK ((seats_booked > 0))
);


--
-- Name: ride_stops; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ride_stops (
    id uuid NOT NULL,
    ride_id uuid NOT NULL,
    location public.geometry(Point,4326) NOT NULL,
    address character varying,
    order_index integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone
);


--
-- Name: rides; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rides (
    id uuid NOT NULL,
    driver_id uuid NOT NULL,
    vehicle_id uuid NOT NULL,
    origin public.geometry(Point,4326) NOT NULL,
    destination public.geometry(Point,4326) NOT NULL,
    origin_address character varying NOT NULL,
    destination_address character varying NOT NULL,
    departure_time timestamp with time zone NOT NULL,
    total_seats integer NOT NULL,
    available_seats integer NOT NULL,
    price_per_seat numeric(10,2) NOT NULL,
    gender_restriction character varying(20) DEFAULT NULL::character varying,
    status character varying DEFAULT 'created'::character varying NOT NULL,
    is_negotiable boolean DEFAULT false NOT NULL,
    has_fair_price_badge boolean DEFAULT false NOT NULL,
    has_ac boolean DEFAULT false NOT NULL,
    has_luggage boolean DEFAULT false NOT NULL,
    takes_parcels boolean DEFAULT false NOT NULL,
    no_smoking boolean DEFAULT false NOT NULL,
    no_pets boolean DEFAULT false NOT NULL,
    driving_speed character varying(50) DEFAULT 'normal'::character varying,
    max_two_back boolean DEFAULT false NOT NULL,
    pickup_point character varying(255) DEFAULT ''::character varying NOT NULL,
    vehicle_data jsonb,
    route_polygon public.geometry(Polygon,4326),
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    deleted_at timestamp with time zone,
    cancel_reason text,
    cancelled_at timestamp with time zone,
    CONSTRAINT rides_available_seats_check CHECK ((available_seats >= 0)),
    CONSTRAINT rides_total_seats_check CHECK ((total_seats > 0))
);


--
-- Name: route_subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.route_subscriptions (
    id uuid NOT NULL,
    passenger_id uuid NOT NULL,
    origin_lat double precision NOT NULL,
    origin_lon double precision NOT NULL,
    dest_lat double precision NOT NULL,
    dest_lon double precision NOT NULL,
    departure_date date,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: sent_reminders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sent_reminders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    ride_id uuid NOT NULL,
    user_id uuid NOT NULL,
    kind character varying(32) NOT NULL,
    sent_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Data for Name: bookings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.bookings (id, ride_id, passenger_id, seats_booked, total_price, is_full_car_buyout, offered_price, status, pickup_preference, created_at, updated_at, deleted_at) FROM stdin;
9e998a28-01c2-4c23-8bb1-15d01495ef13	c1492d2c-23e4-4da3-8aa3-8fd32a684379	8bc95322-9fc9-4903-b893-79c37aa115bc	3	1000000.00	t	1000000.00	rejected	\N	2026-08-19 02:08:48.441908+00	2026-08-19 02:39:46.47728+00	\N
5f1c5558-81bb-4b8e-b849-17656960e368	c1492d2c-23e4-4da3-8aa3-8fd32a684379	8bc95322-9fc9-4903-b893-79c37aa115bc	3	300000.00	t	\N	pending	\N	2026-08-19 02:40:16.982687+00	2026-08-19 02:40:16.982687+00	\N
\.


--
-- Data for Name: ride_stops; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ride_stops (id, ride_id, location, address, order_index, created_at, updated_at, deleted_at) FROM stdin;
0f9aaed1-b3c5-43c9-8220-a6ec28e0f963	29ab008d-cd8f-44a3-801d-3916ca71e7b9	0101000020E61000006519E258175751400BB5A679C7814440	Nurafshon	0	2026-08-20 00:57:07.244238+00	2026-08-20 00:57:07.244238+00	\N
\.


--
-- Data for Name: rides; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.rides (id, driver_id, vehicle_id, origin, destination, origin_address, destination_address, departure_time, total_seats, available_seats, price_per_seat, gender_restriction, status, is_negotiable, has_fair_price_badge, has_ac, has_luggage, takes_parcels, no_smoking, no_pets, driving_speed, max_two_back, pickup_point, vehicle_data, route_polygon, created_at, updated_at, deleted_at, cancel_reason, cancelled_at) FROM stdin;
c76073bd-667e-4e7c-a5f5-18d84ab1e681	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	c6ced4f1-1fb7-4b41-851a-43b653c4d7bd	0101000020E610000072175B3AF1F35140304EF8732C894440	0101000020E6100000371F3240FD5651403C399E2A3A844440	Chortoq	O&#39;rtachirchiq	2026-08-20 10:00:00+00	3	3	100000.00		draft	t	t	f	t	t	f	f	normal	f		{"id": "c6ced4f1-1fb7-4b41-851a-43b653c4d7bd", "year": 0, "brand": "Chevrolet", "color": "", "model": "Malibu", "license_plate": "01R125MK", "moderation_status": "approved"}	0103000020E61000000100000023000000DAA90A351F5751401BA66E2B777E4440D592F1AC6056514029A93F75897E444044161426A8555140EBD5B4B0D37E44402CB88CB4FC545140A7BCEF04537F444063F2B6EC645451409874878F0280444063DACDA2E653514078F83E94DB804440A4F0B9B186535140FF37DEBED58144406CE53ACB48535140FB0CB674E7824440669F46532F5351409A99BD32068444406CC410483B535140D6A1B9F426854440071BB8376C535140E24F81A13E86444055D71444C05351407AB73B784287444045A69834345451407E58617A28884440C239A495C3545140F67266CEE788444067012CE468555140FD303617798944400BC604C41D565140D1B111BDD68944408632BF3EDB565140593CF524FD894440938CB1AAEAF351403785264AEF8E444004355405A9F45140FF1305DDD48E44400CD5C64B60F55140293D75BD828E444089FDE46F09F6514000612415FC8D444008C1ABEF9DF65140EE3A4E13468D444090498E1518F7514085835AB9678C44409A98B83073F7514007DD8295698B4440FC5D13C3ABF75140014B3C6E558A44407F8349A3BFF751403EC5B4E13589444076108E11AEF75140DFA527FD158844406E5069BE77F7514009DD16D000874440A69160C31EF7514041E598FF00864440C789D68DA6F651403716E45D20854440E63E06BD13F651400570108A67844440BFC072F46BF551400C01B49BDD8344409E678FA4B5F45140D57C87DD87834440C6E2BBCBF7F351405C4CAD9969834440DAA90A351F5751401BA66E2B777E4440	2026-08-18 01:26:23.939508+00	2026-08-18 01:26:23.952258+00	\N	\N	\N
c1492d2c-23e4-4da3-8aa3-8fd32a684379	dd5dd41d-c0b1-4650-81f8-e6d9c0f5789a	c6ced4f1-1fb7-4b41-851a-43b653c4d7bd	0101000020E6100000371F3240FD5651403C399E2A3A844440	0101000020E610000072175B3AF1F35140304EF8732C894440	O&#39;rtachirchiq	Chortoq	2026-08-19 08:30:00+00	3	0	100000.00		completed	t	t	f	t	t	f	f	normal	f		{"id": "c6ced4f1-1fb7-4b41-851a-43b653c4d7bd", "year": 0, "brand": "Chevrolet", "color": "", "model": "Malibu", "license_plate": "01R125MK", "moderation_status": "approved"}	0103000020E61000000100000023000000938CB1AAEAF351403785264AEF8E444004355405A9F45140FF1305DDD48E44400CD5C64B60F55140293D75BD828E444089FDE46F09F6514000612415FC8D444008C1ABEF9DF65140EE3A4E13468D444090498E1518F7514085835AB9678C44409A98B83073F7514007DD8295698B4440FC5D13C3ABF75140014B3C6E558A44407F8349A3BFF751403EC5B4E13589444076108E11AEF75140DFA527FD158844406E5069BE77F7514009DD16D000874440A69160C31EF7514041E598FF00864440C789D68DA6F651403716E45D20854440E63E06BD13F651400570108A67844440BFC072F46BF551400C01B49BDD8344409E678FA4B5F45140D57C87DD87834440C6E2BBCBF7F351405C4CAD9969834440DAA90A351F5751401BA66E2B777E4440D592F1AC6056514029A93F75897E444044161426A8555140EBD5B4B0D37E44402CB88CB4FC545140A7BCEF04537F444063F2B6EC645451409874878F0280444063DACDA2E653514078F83E94DB804440A4F0B9B186535140FF37DEBED58144406CE53ACB48535140FB0CB674E7824440669F46532F5351409A99BD32068444406CC410483B535140D6A1B9F426854440071BB8376C535140E24F81A13E86444055D71444C05351407AB73B784287444045A69834345451407E58617A28884440C239A495C3545140F67266CEE788444067012CE468555140FD303617798944400BC604C41D565140D1B111BDD68944408632BF3EDB565140593CF524FD894440938CB1AAEAF351403785264AEF8E4440	2026-08-18 01:25:59.497462+00	2026-08-20 00:49:36.377599+00	\N	\N	\N
29ab008d-cd8f-44a3-801d-3916ca71e7b9	67a1ef47-5ea2-4ed6-af06-5e82a13678dc	d336419a-6ef2-4fdc-b898-f141e295d4b2	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E610000029ED0DBE30E951408D28ED0DBE804440	Toshkent	Namangan	2026-08-21 17:00:00+00	4	4	50000.00		draft	t	t	f	t	t	f	f	normal	f	qoylu	\N	0103000020E6100000010000002C0000002A860064E759514086A5C64794874440D698E90E42E9514042C48ED380864440F2AD77C0FFE951408E4224595F86444076891979B5EA514071618372068644401436123A5CEB5140A2A91E8C79854440D9D50399EDEB5140AE08A912BE844440BD6055FF63EC514035B6543DDB83444068F63DE1BAEC514083C867C6D982444096B855EAEEEC51406415F694C3814440290EFC1DFEEC51406F3E215BA380444048B162EAE7EC5140B555AD2C847F4440F0EB952DADEC5140B4B2FE11717E44407DCC662C50EC5140D67AB29B747D44408625A67BD4EB514000ECF67A987C444096B5A1DC3EEB5140A41A9422E57B4440887E4B0E95EA5140556A3C74617B44403E98DA94DDE951401492417D127B444022EF097A1FE95140E46A3445FB7A4440616603851F575140A2FABDA1037C44403364B7EB6D565140BA2B55151A7C44403734C8FCC155514069DC9D33617C444005845E7821555140D9EEBF9BD67C44405B2973BD91545140C0D47660767D4440B388FD9B17545140B40A83293B7E4440FB31CB2BB75351404FFC37611E7F44405D935DA973535140E963A96C1880444000ADBBD9BD4B514008D1EE8B20A644409970B9C1974B514033A3F7183CA744404EE91512974B5140C9F17A685DA84440A71C65D5BB4B5140DFC6415C79A94440CD1987A5044C5140475AB40985AA4440DAAAE0B86E4C5140BF2C692576AB4440BB1A5DFDF54C5140E083BA6843AC444023C53F40954D5140F2026FEDE4AC444029C54661464E514071F1DF7B54AD444038FA278F024F5140C53D8BC88DAD444045BA0C8BC24F5140289FA39E8EAD44402B2564F07E5051402BE9EFF556AD4440B247347E30515140EEC81BF3E8AC44409EB8015FD0515140CED776D248AC44409B02756C5852514039C0F9BD7CAB44408221146CC35251409EAE3A908CAA444076FEB1420D535140C985B98681A944402A860064E759514086A5C64794874440	2026-08-20 00:57:07.244238+00	2026-08-20 00:57:07.251944+00	\N	\N	\N
1114d90b-c16e-4aab-b309-0c531e9406a9	60513e3b-5f2a-4b2c-a37b-6e726c089d00	ffffffff-ffff-4fff-8fff-ffffffffffff	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E6100000C364AA6054F251405396218E75314440	Toshkent	Farg%27ona	2026-08-25 03:00:00+00	3	3	100000.00		cancelled	f	t	f	f	f	f	f	\N	f		\N	\N	2026-08-21 14:44:31.663121+00	2026-08-21 14:45:10.805924+00	\N	\N	\N
301fed87-63a0-4119-ab26-e2a3b7edc92e	60513e3b-5f2a-4b2c-a37b-6e726c089d00	ffffffff-ffff-4fff-8fff-ffffffffffff	0101000020E610000043AA285E654F514018B49080D1A74440	0101000020E6100000C364AA6054F251405396218E75314440	Toshkent	Fargona	2026-08-26 03:00:00+00	3	3	100000.00		cancelled	f	t	f	f	f	f	f	\N	f		\N	\N	2026-08-21 20:53:34.326177+00	2026-08-21 20:53:51.2205+00	\N	\N	\N
\.


--
-- Data for Name: route_subscriptions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.route_subscriptions (id, passenger_id, origin_lat, origin_lon, dest_lat, dest_lon, departure_date, is_active, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
1	f
\.


--
-- Data for Name: sent_reminders; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sent_reminders (id, ride_id, user_id, kind, sent_at) FROM stdin;
\.


--
-- Data for Name: spatial_ref_sys; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.spatial_ref_sys (srid, auth_name, auth_srid, srtext, proj4text) FROM stdin;
\.


--
-- Name: bookings bookings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_pkey PRIMARY KEY (id);


--
-- Name: ride_stops ride_stops_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ride_stops
    ADD CONSTRAINT ride_stops_pkey PRIMARY KEY (id);


--
-- Name: rides rides_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rides
    ADD CONSTRAINT rides_pkey PRIMARY KEY (id);


--
-- Name: route_subscriptions route_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.route_subscriptions
    ADD CONSTRAINT route_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: sent_reminders sent_reminders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sent_reminders
    ADD CONSTRAINT sent_reminders_pkey PRIMARY KEY (id);


--
-- Name: sent_reminders unique_ride_user_kind; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sent_reminders
    ADD CONSTRAINT unique_ride_user_kind UNIQUE (ride_id, user_id, kind);


--
-- Name: idx_bookings_passenger_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_bookings_passenger_id ON public.bookings USING btree (passenger_id);


--
-- Name: idx_bookings_ride_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_bookings_ride_id ON public.bookings USING btree (ride_id);


--
-- Name: idx_ride_stops_location; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ride_stops_location ON public.ride_stops USING gist (location);


--
-- Name: idx_rides_destination; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_rides_destination ON public.rides USING gist (destination);


--
-- Name: idx_rides_origin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_rides_origin ON public.rides USING gist (origin);


--
-- Name: idx_rides_route_polygon; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_rides_route_polygon ON public.rides USING gist (route_polygon);


--
-- Name: idx_rides_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_rides_status ON public.rides USING btree (status);


--
-- Name: idx_route_subs_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_route_subs_active ON public.route_subscriptions USING btree (is_active) WHERE (is_active = true);


--
-- Name: idx_route_subs_passenger; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_route_subs_passenger ON public.route_subscriptions USING btree (passenger_id);


--
-- Name: bookings bookings_ride_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bookings
    ADD CONSTRAINT bookings_ride_id_fkey FOREIGN KEY (ride_id) REFERENCES public.rides(id) ON DELETE CASCADE;


--
-- Name: ride_stops ride_stops_ride_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ride_stops
    ADD CONSTRAINT ride_stops_ride_id_fkey FOREIGN KEY (ride_id) REFERENCES public.rides(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

