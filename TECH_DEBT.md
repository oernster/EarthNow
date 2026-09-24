# EarthNow: Technical Debt

A standing reference to what is still open, what is deliberately left and what only looks like debt. Most of EarthNow's open items are known limits rather than internal debt: small places where the product gives a wrong or weak answer for a rare input, parked by the owner to be dealt with later rather than fixed on sight. Each one says what fixing it would change, so none is mistaken for a behaviour-preserving tidy. Scope is the whole repository read against `REQUIREMENTS.md` and the structural tests in `tests/structural`.

---

## 1. A point on the Ross Ice Shelf reads "At sea"

`internal/infrastructure/geo/geo.go` words a point as at sea when the Natural Earth country polygons contain it nowhere. Antarctic ice shelves are not land in that data set, so an event on the Ross Ice Shelf reads "At sea; ..." although it is on ice (seen on a sea-ice event in the session that built the place line, 2026-09-23).

Fixing it means a second source of land for Antarctica (Natural Earth publishes ice shelves as their own layer) or a wording that does not claim sea where the data only says "not a country"; either changes the place line.

## 2. The GDACS latitude-first rule rests on one week's polygons

EONET passes GDACS flood polygons on with each vertex as [lat, lng], against GeoJSON's [lng, lat], while GDACS points arrive in GeoJSON order (all 14 GDACS polygons of the week to 2026-09-23, measured; REQUIREMENTS amendment 12). `internal/infrastructure/providers/eonet/eonet.go` reads every GDACS polygon latitude first on that evidence.

**Blocked on time.** This is a verification gap rather than a known defect. If EONET or GDACS corrects the order, every GDACS polygon will be drawn swapped, flood markers moving to impossible places (the week that found the rule put Honduras in Antarctica). The item closes when a later week's GDACS polygons are measured again and still arrive latitude first. Replacing the rule with one that decides per polygon would close it too.

## 3. Appendix C's traceability test does not exist yet

REQUIREMENTS.md Appendix C promises a structural test that lists every Must and fails when no test names it. Measured on 2026-09-24: 64 of the 134 Musts are named by no test; 31 of those are marked T, the rest D, I or a timing measurement. The count is a little high, since a test named "FR-DON-001 and 005" covers 005 without spelling the full ID.

Closing it means naming (or writing) a test for each T requirement, listing the D and I ones in TESTING.md's "Checked by a person" table, then adding the test that holds both lists to the document. It changes no behaviour; it is a session of work rather than a line.

---

## Looks like debt, not worth touching

- **The setup package's coverage floor sits at 59.9%.** What it leaves uncovered acts on the machine itself: the registry writes, shortcut creation through the Windows Script Host and the process work. A test must not change the machine it runs on. `test.ps1` names the gap beside the floor.

## Not debt (do not "fix" these)

- **The setup package reads 61.4% on some machines and 59.9% on others.** The installed-version read runs three statements further where EarthNow is installed (121 of 197 against 118, measured). The floor is the figure for a machine without it; raising the floor to 61.4 would fail the gate on any machine that has never installed the product.
- **An event dated more than `window.ClockSkew` ahead of the clock is not shown.** GDACS published a flood alert dated days ahead (EONET_24511, seen on 2026-09-23, carried 2026-10-04 and GDACS gave 5 to 7 October). Why the source dates it so is not known; whatever it is, it has not happened yet, so FR-TW-002 leaves it out until its time arrives.
