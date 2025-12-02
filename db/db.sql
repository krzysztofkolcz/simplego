CREATE DATABASE simplego;

CREATE TABLE public.products (
  id serial PRIMARY KEY,
  price int
);

INSERT INTO products (price) VALUES (100);