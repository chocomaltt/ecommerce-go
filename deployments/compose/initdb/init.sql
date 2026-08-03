-- Local dev only. Runs once on first postgres init.
-- One database per service + Kratos, owned by a shared dev user.
CREATE USER ecommerce WITH PASSWORD 'ecommerce';

CREATE DATABASE catalog OWNER ecommerce;
CREATE DATABASE inventory OWNER ecommerce;
CREATE DATABASE orders OWNER ecommerce;
CREATE DATABASE payments OWNER ecommerce;
CREATE DATABASE notifications OWNER ecommerce;
CREATE DATABASE kratos OWNER ecommerce;
