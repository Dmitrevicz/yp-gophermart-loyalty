
-- orders table
CREATE TABLE IF NOT EXISTS orders(
   id VARCHAR(100) PRIMARY KEY, -- length of 100 chars is set as a reasonable limit to prevent gigabytes of junk being saved
   user_id bigint NOT NULL,
   uploaded_at timestamptz NOT NULL DEFAULT now(),
   status VARCHAR(50) NOT NULL DEFAULT '',
   accrual numeric(20,4) NOT NULL DEFAULT 0,
   processed_at timestamptz NULL,
   CONSTRAINT fk_user_id
      FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- users' loyalty points balance
CREATE TABLE IF NOT EXISTS loyalty_points(
   user_id bigint NOT NULL PRIMARY KEY,
   balance numeric(20,4) NOT NULL DEFAULT 0,
   updated timestamptz NOT NULL DEFAULT now(),
   CONSTRAINT fk_user_id
      FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- withdrawals history
CREATE TABLE IF NOT EXISTS withdrawals(
   user_id bigint NOT NULL,
   id UUID NOT NULL PRIMARY KEY,
   -- order_number specs: по ТЗ ожидается всего лишь гипотетический номер нового заказа пользователя, 
   -- поэтому не добавляю на этот столбец внешний ключ к таблице заказов
   order_number VARCHAR(100) NOT NULL DEFAULT '',
   value numeric(20,4) NOT NULL DEFAULT 0,
   processed_at timestamptz NOT NULL DEFAULT now(),
   CONSTRAINT fk_user_id
      FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
