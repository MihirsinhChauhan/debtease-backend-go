from loan_portfolio_engine import Loan, generate_portfolio_schedule
from stress_engine import simulate_stress


def test_no_debts_resilient():
    outcome = simulate_stress(
        income=60000.0,
        fixed_expenses=10000.0,
        schedule={},
        loans=[],
        month=1,
    )

    assert outcome.stress_level == "RESILIENT"
    assert outcome.failed_debt_categories == []


def test_income_shock_can_be_fragile():
    # Unsecured personal loan; income shock + rate shock should push buffer negative.
    loans = [
        Loan(
            loan_id="PL",
            principal=5_00_000,
            annual_rate=13.0,
            tenure_months=60,
            instrument_type="personal_loan",
            secured=False,
        )
    ]
    schedule = generate_portfolio_schedule(loans)

    outcome = simulate_stress(
        income=60_000.0,
        fixed_expenses=10_000.0,
        schedule=schedule,
        loans=loans,
        month=1,
    )

    # Depending on precise EMI math and thresholds, this scenario should at least
    # not be resilient. It should be either sensitive or fragile with low buffer.
    assert outcome.stress_level in ("SENSITIVE", "FRAGILE")
    assert any(
        d in outcome.key_drivers
        for d in ("Low flexibility buffer", "Thin flexibility buffer")
    )

