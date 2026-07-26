# API ↔ FE Wiring Checklists

Đã tách thành 3 file theo app / lớp:

| # | File | Phạm vi |
| --- | --- | --- |
| 00 | [api-wiring/00-foundation.md](./api-wiring/00-foundation.md) | Policy httpOnly JWT proxy, quy trình wire, shared packages, cross-cutting, DoD |
| 01 | [api-wiring/01-admin-workspace.md](./api-wiring/01-admin-workspace.md) | `apps/admin` — org ops, connector registry, admin shell |
| 02 | [api-wiring/02-user-workspace.md](./api-wiring/02-user-workspace.md) | `apps/web` — ma trận endpoint, gaps, feature screens |

Thứ tự làm việc khuyến nghị: **00 → 02** (user workspace đang active) → **01** (admin phase sau).
