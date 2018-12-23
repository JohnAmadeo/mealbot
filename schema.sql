DROP TABLE rounds;
DROP TABLE pairs;
DROP TABLE members;
DROP TABLE organizations;

CREATE TABLE organizations (
    name VARCHAR PRIMARY KEY,
    admin VARCHAR UNIQUE NOT NULL CHECK(length(admin) > 0),
    cross_match_criteria VARCHAR
);

CREATE TABLE members (
    organization VARCHAR REFERENCES organizations(name),
    email VARCHAR PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL CHECK(length(name) > 0),
    metadata JSONB,
    pair_counts JSONB NOT NULL
);

CREATE TABLE pairs (
    organization VARCHAR REFERENCES organizations(name),
    id1 VARCHAR REFERENCES members(email),
    id2 VARCHAR REFERENCES members(email),
    round INTEGER NOT NULL CHECK(round > 0)
);

CREATE TABLE rounds (
    organization VARCHAR REFERENCES organizations(name),
    scheduled_date DATE NOT NULL
);