CREATE TYPE account_type AS ENUM ('FREE', 'PREMIUM');
CREATE TYPE transaction_status AS ENUM ('UNPAID', 'PAID', 'FAILED', 'EXPIRED', 'REFUND');
CREATE TYPE indonesia_time_zone AS ENUM ('Asia/Jakarta', 'Asia/Makassar', 'Asia/Jayapura');
CREATE TYPE prayer_status AS ENUM ('ON_TIME', 'LATE', 'MISSED');

CREATE TABLE "user" (
  id VARCHAR(255),
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  phone_number VARCHAR(255) UNIQUE,
  phone_verified BOOLEAN DEFAULT FALSE NOT NULL,
  account_type account_type DEFAULT 'FREE' NOT NULL,
  time_zone indonesia_time_zone,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

  PRIMARY KEY (id)
);

CREATE TABLE coupon (
  code VARCHAR(255),
  influencer_username VARCHAR(255) NOT NULL,
  quota SMALLINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_at TIMESTAMPTZ,
  
  PRIMARY KEY (code)
);

CREATE TABLE subscription_plan (
  id UUID DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  price INT NOT NULL,
  duration_in_months SMALLINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  deleted_at TIMESTAMPTZ,

  PRIMARY KEY (id),
  UNIQUE (name, duration_in_months)
);

CREATE TABLE transaction (
  id UUID,
  user_id VARCHAR(255) NOT NULL,
  subscription_plan_id UUID NOT NULL,
  ref_id VARCHAR(255) NOT NULL,
  coupon_code VARCHAR(255),
  payment_method VARCHAR(255) NOT NULL,
  qr_url VARCHAR(255) NOT NULL,
  status transaction_status DEFAULT 'UNPAID' NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  paid_at TIMESTAMPTZ,
  expired_at TIMESTAMPTZ NOT NULL,

  PRIMARY KEY (id),
  
  CONSTRAINT fk_user_transaction
    FOREIGN KEY (user_id)
    REFERENCES "user"(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

  CONSTRAINT fk_subscription_plan
    FOREIGN KEY (subscription_plan_id)
    REFERENCES subscription_plan(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

  CONSTRAINT fk_coupon
    FOREIGN KEY (coupon_code)
    REFERENCES coupon(code)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);

CREATE TABLE prayer (
  id UUID DEFAULT gen_random_uuid(),
  user_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  status prayer_status,
  year SMALLINT NOT NULL,
  month SMALLINT NOT NULL,
  day SMALLINT NOT NULL,

  PRIMARY KEY (id),

  CONSTRAINT fk_user_prayer
    FOREIGN KEY (user_id)
    REFERENCES "user"(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE,

  CONSTRAINT unique_prayer UNIQUE (user_id, name, year, month, day)
);

CREATE TABLE task (
  id UUID DEFAULT gen_random_uuid(),
  user_id VARCHAR(255) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  checked BOOLEAN DEFAULT FALSE NOT NULL,

  PRIMARY KEY (id),

  CONSTRAINT fk_user_task
    FOREIGN KEY (user_id)
    REFERENCES "user"(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
);
