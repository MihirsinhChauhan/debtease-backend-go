-- name: GetPortfolioRollupForUser :one
-- Aligns portfolio EMI/balance aggregates with dashboard debt filters (positive outstanding).
SELECT
    COALESCE(SUM(outstanding_balance + accrued_interest), 0)::numeric AS total_debt_balance,
    COALESCE(SUM(COALESCE(emi_amount, min_payment)), 0)::numeric AS total_monthly_emi,
    COUNT(*)::bigint AS active_debt_count
FROM debts
WHERE user_id = $1::uuid
  AND outstanding_balance > 0;
