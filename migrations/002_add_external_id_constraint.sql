ALTER table transactions
ADD CONSTRAINT unique_external_id UNIQUE (external_id);