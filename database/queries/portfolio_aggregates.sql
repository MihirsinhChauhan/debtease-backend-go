-- name: GetPortfolioRollupForUser :one
-- Aligns portfolio obligation/balance aggregates with dashboard debt filters (positive outstanding).
SELECT
    COALESCE(SUM(outstanding_balance + accrued_interest), 0)::numeric AS total_debt_balance,
    COALESCE(SUM(monthly_obligation), 0)::numeric AS total_monthly_emi,
    COUNT(*)::bigint AS active_debt_count,
    COALESCE(SUM(CASE WHEN NOT is_secured THEN monthly_obligation ELSE 0::numeric END), 0)::numeric AS unsecured_monthly_obligation,
    COALESCE(SUM(CASE WHEN is_revolving THEN monthly_obligation ELSE 0::numeric END), 0)::numeric AS revolving_monthly_obligation,
    COALESCE(SUM(CASE WHEN is_short_term THEN monthly_obligation ELSE 0::numeric END), 0)::numeric AS short_term_obligations
FROM debts
WHERE user_id = $1::uuid
  AND outstanding_balance > 0;
