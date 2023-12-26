ALTER TABLE loyalty_points
ADD CONSTRAINT balance_not_negative_check
CHECK (balance >= 0);
