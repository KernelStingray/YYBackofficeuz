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

ALTER TABLE IF EXISTS ONLY public.feedbacks DROP CONSTRAINT IF EXISTS feedbacks_pkey;
DROP TABLE IF EXISTS public.feedbacks;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: feedbacks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feedbacks (
    id uuid NOT NULL,
    ride_id uuid NOT NULL,
    reviewer_id uuid NOT NULL,
    reviewee_id uuid NOT NULL,
    score integer NOT NULL,
    comment text,
    created_at timestamp without time zone NOT NULL
);


--
-- Data for Name: feedbacks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.feedbacks (id, ride_id, reviewer_id, reviewee_id, score, comment, created_at) FROM stdin;
\.


--
-- Name: feedbacks feedbacks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feedbacks
    ADD CONSTRAINT feedbacks_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

