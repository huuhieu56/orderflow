ALTER TABLE orders
    ADD CONSTRAINT orders_total_amount_check CHECK (total_amount >= 0),
    ADD CONSTRAINT orders_status_check CHECK (status IN ('pending', 'cancelled'));

ALTER TABLE order_items
    ADD CONSTRAINT order_items_unit_price_check CHECK (unit_price > 0),
    ADD CONSTRAINT order_items_quantity_check CHECK (quantity > 0),
    ADD CONSTRAINT order_items_subtotal_check CHECK (subtotal > 0);
