# EarthNow: Technical Debt

A standing reference to what is still open, what is deliberately left and what only looks like debt. Most of EarthNow's open items are known limits rather than internal debt: small places where the product gives a wrong or weak answer for a rare input, parked by the owner to be dealt with later rather than fixed on sight. Each one says what fixing it would change, so none is mistaken for a behaviour-preserving tidy. Scope is the whole repository read against `REQUIREMENTS.md` and the structural tests in `tests/structural`.

---

There is no open technical debt.

---

## Looks like debt, not worth touching

- **The setup package's coverage floor sits at 59.9%.** What it leaves uncovered acts on the machine itself: the registry writes, shortcut creation through the Windows Script Host and the process work. A test must not change the machine it runs on. `test.ps1` names the gap beside the floor.

## Not debt (do not "fix" these)

- **The setup package reads 61.4% on some machines and 59.9% on others.** The installed-version read runs three statements further where EarthNow is installed (121 of 197 against 118, measured). The floor is the figure for a machine without it; raising the floor to 61.4 would fail the gate on any machine that has never installed the product.
- **An event dated more than `window.ClockSkew` ahead of the clock is not shown.** GDACS published a flood alert dated days ahead (EONET_24511, seen on 2026-09-23, carried 2026-10-04 and GDACS gave 5 to 7 October). Why the source dates it so is not known; whatever it is, it has not happened yet, so FR-TW-002 leaves it out until its time arrives.
