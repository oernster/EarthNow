# EarthNow: Technical Debt

A standing reference to what is still open, what is deliberately left and what only looks like debt. An open item says what fixing it would change, so none is mistaken for a behaviour-preserving tidy. Scope is the whole repository read against `REQUIREMENTS.md` and the structural tests in `tests/structural`.

---

There is no open technical debt.

---

## Looks like debt, not worth touching

- **The setup package's coverage floor sits at 59.9%.** What it leaves uncovered acts on the machine itself: the registry writes, shortcut creation through the Windows Script Host and the process work. A test must not change the machine it runs on. `test.ps1` names the gap beside the floor. It reads 61.4% where EarthNow is installed, since the installed-version read then runs three statements further (121 of 197 against 118, measured); the floor is the figure for a machine without it, so raising it would fail the gate on any machine that has never installed the product.
- **The settings file reads and writes on its own.** `internal/infrastructure/settings` carries its own capped read and write-then-rename, much as `cache` shares one reader and one writer between the providers' files, the cloud image and the burnt-area days. Folding settings into those helpers would change its error wording and its temporary file's name for no behaviour the reader sees; the owner judged it fine as it stands (2026-09-24).

## Not debt (do not "fix" these)

- **An event dated more than `window.ClockSkew` ahead of the clock is not shown.** GDACS published a flood alert dated days ahead (EONET_24511, seen on 2026-09-23, dated 2026-10-04). Why the source dates it so is not known; whatever it is, it has not happened yet, so FR-TW-002 leaves it out until its time arrives.
