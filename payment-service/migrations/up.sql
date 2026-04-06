CREATE TABLE payments (
                          id VARCHAR(36) PRIMARY KEY,
                          order_id VARCHAR(36),
                          transaction_id VARCHAR(36),
                          amount BIGINT,
                          status VARCHAR(50)
);