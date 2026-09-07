ALTER TABLE products
    ADD CONSTRAINT products_price_check CHECK (price > 0),
    ADD CONSTRAINT products_stock_check CHECK (stock >= 0),
    ADD CONSTRAINT products_status_check CHECK (status IN ('active', 'inactive'));
