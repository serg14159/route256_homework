-- name: GetAvailableStockBySKU :one
SELECT total_count - reserved AS available
FROM stocks
WHERE sku = $1;

-- name: ReserveItems :exec
UPDATE stocks
SET reserved = reserved + $2
WHERE sku = $1;

-- name: RemoveReservedItems :exec
UPDATE stocks
SET reserved = reserved - $2, total_count = total_count - $2
WHERE sku = $1;

-- name: CancelReservedItems :exec
UPDATE stocks
SET reserved = reserved - $2
WHERE sku = $1;