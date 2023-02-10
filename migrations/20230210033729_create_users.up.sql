CREATE TABLE users(login TEXT NOT NULL PRIMARY KEY,password TEXT NOT NULL, balance INT DEFAULT 0 CHECK (balance >= 0));
