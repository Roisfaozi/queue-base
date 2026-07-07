# QMS Feature Map Addendum

Latest design priority:

- `documentation/New Design Document — QMS MVP Operatio.md`

This feature map adds QMS domain slices on top of the starter roadmap model.

## QMS slices

- tenant and branch foundation
- qms_clients and credential binding
- operator counter assignments
- queue master
- queue_journeys
- visit_journeys
- scanner
- typed settings inheritance via branch_service scope

## Rule

Forward belongs to journey history, not to a second master queue row.

## QMS Test Matrix Rule

Each slice should define tests in four buckets:

- positive
- negative
- edge
- vulnerability/security
