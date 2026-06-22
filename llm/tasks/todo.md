# Todo

Status:

- queue rebuild framing complete
- feature inventory extracted
- module map drafted
- active coding not started yet

Current execution plan:

1. finalize bounded-context names for queue, forward, scanner, service-point, and patient/customer
2. design migrations and API contracts for queue core
3. implement queue registration and counter locking first
4. implement forward orchestration and pharmacy validation second
5. implement scanner auth/check-in third
6. wire integration adapters and consumer sync last

Completed in this pass:

- extracted complete requirement set from `documentation/task-overview.md`
- extracted architecture flow set from `documentation/project-diagram.md`
- mapped all requested features into repo-native module candidates
- wrote `llm/research/queue-management-domain-map.md`
- wrote `llm/plans/roadmap/queue-management-feature-map.md`
- kept legacy `code_context.txt` as evidence only, not runtime truth

Immediate next step for implementation session:

- choose final primary module naming and start schema design for queue core
