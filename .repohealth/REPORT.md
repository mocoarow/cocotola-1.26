# Repository health

Snapshot of 37eda81def47e75d03a5c45789bd5b750fe17153 committed at 2026-11-24T00:05:03+09:00, taken at 2026-11-24T00:12:09+09:00 (repohealth v0.0.0-20261123133459-d188564a8b95, Biome 2.4.10).

## Areas

| Area | Go lines | Go test lines | Go functions | Go cyclo max | Go cognit sum | Go cognit max | Go over limit | TS lines | TS test lines | TS functions (cognit ≥ 2) | TS cognit sum | TS cognit max | TS over limit | Coverage |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| app | 213 | 75 | 4 | 9 | 15 | 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 77.8% |
| audio-generator | 953 | 1027 | 31 | 9 | 85 | 10 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 72.3% |
| auth | 10792 | 17475 | 608 | 12 | 786 | 17 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 83.9% |
| empty | 179 | 0 | 8 | 5 | 13 | 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | - |
| init | 2268 | 3715 | 82 | 10 | 266 | 15 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 86.4% |
| lib | 2049 | 4668 | 91 | 8 | 155 | 15 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 74.7% |
| question | 10859 | 19271 | 497 | 11 | 961 | 13 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 85.5% |
| test | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 457 | 1799 | 2 | 4 | 2 | 0 | - |
| web | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 9025 | 17032 | 206 | 785 | 15 | 0 | - |

## Trends

### Coverage by area

| Committed | Rev | app | audio-generator | auth | empty | init | lib | question | test | web |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-24T00:05:03+09:00 | 37eda81def47e75d03a5c45789bd5b750fe17153 | 77.8% | 72.3% | 83.9% | - | 86.4% | 74.7% | 85.5% | - | - |

### Go cognitive complexity (sum)

| Committed | Rev | app | audio-generator | auth | empty | init | lib | question | test | web |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-24T00:05:03+09:00 | 37eda81def47e75d03a5c45789bd5b750fe17153 | 15 | 85 | 786 | 13 | 266 | 155 | 961 | 0 | 0 |

### TS cognitive complexity (sum)

| Committed | Rev | app | audio-generator | auth | empty | init | lib | question | test | web |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-24T00:05:03+09:00 | 37eda81def47e75d03a5c45789bd5b750fe17153 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 4 | 785 |

## Cross-area changes

### Months

| Month | PRs | Cross-area | Cross-area (backend) |
| --- | --- | --- | --- |
| 2026-03 | 15 | 8 | 0 |
| 2026-04 | 60 | 34 | 17 |
| 2026-05 | 20 | 10 | 3 |
| 2026-06 | 22 | 4 | 2 |
| 2026-07 | 11 | 0 | 0 |
| 2026-10 | 65 | 7 | 3 |
| 2026-11 | 81 | 13 | 4 |

### Pairs

| Areas | PRs |
| --- | --- |
| auth + question | 25 |
| auth + init | 22 |
| question + web | 20 |
| auth + lib | 16 |
| lib + question | 15 |
| init + question | 13 |
| app + question | 12 |
| auth + web | 12 |
| app + auth | 10 |
| init + lib | 9 |

## Dependencies

### Area imports

| From | To | Imports |
| --- | --- | --- |
| auth | lib | 37 |
| question | lib | 25 |
| app | lib | 7 |
| empty | lib | 5 |
| init | auth | 5 |
| init | lib | 5 |
| app | question | 4 |
| app | auth | 2 |

### Stability

| Package | Afferent | Efferent | Instability |
| --- | --- | --- | --- |
| `cocotola-audio-generator/domain` | 2 | 0 | 0.000 |
| `cocotola-auth/domain` | 31 | 3 | 0.088 |
| `cocotola-auth/domain/group` | 2 | 2 | 0.500 |
| `cocotola-auth/domain/rbac` | 10 | 1 | 0.091 |
| `cocotola-auth/domain/space` | 4 | 1 | 0.200 |
| `cocotola-auth/domain/token` | 2 | 1 | 0.333 |
| `cocotola-auth/domain/user` | 4 | 1 | 0.200 |
| `cocotola-lib/domain` | 8 | 0 | 0.000 |
| `cocotola-lib/domain/audio` | 1 | 0 | 0.000 |
| `cocotola-lib/domain/idset` | 3 | 0 | 0.000 |
| `cocotola-lib/domain/versioned` | 8 | 0 | 0.000 |
| `cocotola-question/domain` | 19 | 2 | 0.095 |
| `cocotola-question/domain/question` | 6 | 2 | 0.250 |
| `cocotola-question/domain/reference` | 2 | 1 | 0.333 |
| `cocotola-question/domain/study` | 2 | 1 | 0.333 |
| `cocotola-question/domain/workbook` | 6 | 2 | 0.250 |

## Hotspots

| File | Score | Churn | Recent churn | Cognit sum | Cognit max |
| --- | --- | --- | --- | --- | --- |
| `cocotola-init/initialize/initializer.go` | 1040 | 26 | 10 | 40 | 8 |
| `cocotola-web/app/routes/workbooks.$workbookId.tsx` | 918 | 18 | 9 | 51 | 14 |
| `cocotola-question/domain/question/question.go` | 720 | 15 | 9 | 48 | 13 |
| `cocotola-web/app/routes/workbooks.$workbookId.study.tsx` | 630 | 21 | 10 | 30 | 9 |
| `cocotola-web/app/routes/workbooks.index.tsx` | 612 | 18 | 6 | 34 | 11 |
| `cocotola-init/seed/workbook_seeder.go` | 560 | 14 | 10 | 40 | 10 |
| `cocotola-question/gateway/question_repository.go` | 510 | 10 | 4 | 51 | 12 |
| `cocotola-question/domain/question/content.go` | 477 | 9 | 4 | 53 | 12 |
| `cocotola-question/gateway/auth_client.go` | 408 | 17 | 7 | 24 | 5 |
| `cocotola-question/usecase/workbook/create_workbook_command.go` | 384 | 16 | 3 | 24 | 10 |

## Functions

| Lang | Function | File | Line | Cognit | Cyclo |
| --- | --- | --- | --- | --- | --- |
| go | `(*ListSpacesQuery).ListSpaces` | `cocotola-auth/usecase/space/list_spaces_query.go` | 40 | 17 | 12 |
| go | `validateSeeds` | `cocotola-init/seed/loader.go` | 58 | 15 | 8 |
| go | `(CSVWorkbookManifest).validate` | `cocotola-init/seed/manifest.go` | 65 | 15 | 9 |
| go | `Run` | `cocotola-lib/process/multiprocess.go` | 22 | 15 | 8 |
| go | `NewAuthMiddleware` | `cocotola-auth/controller/middleware/auth_middleware.go` | 29 | 13 | 7 |
| go | `(*Question).validateContent` | `cocotola-question/domain/question/question.go` | 129 | 13 | 10 |
| go | `(*ValidateAccessTokenQuery).ValidateAccessToken` | `cocotola-auth/usecase/auth/validate_access_token_query.go` | 49 | 12 | 10 |
| go | `validateChoices` | `cocotola-question/domain/question/content.go` | 201 | 12 | 9 |
| go | `(*QuestionRepository).Create` | `cocotola-question/gateway/question_repository.go` | 181 | 12 | 9 |
| go | `(*CheckHandler).Check` | `cocotola-auth/controller/handler/authz/check_handler.go` | 48 | 11 | 12 |
| ts | `(anonymous)` | `cocotola-web/app/hooks/use-study-resume.ts` | 50 | 15 | - |
| ts | `WorkbookDetail` | `cocotola-web/app/routes/workbooks.$workbookId.tsx` | 252 | 14 | - |
| ts | `keyAction` | `cocotola-web/app/components/study/question-types/ordering/keyboard.ts` | 101 | 12 | - |
| ts | `loader` | `cocotola-web/app/routes/auth.callback.tsx` | 15 | 12 | - |
| ts | `AddQuestionSection` | `cocotola-web/app/components/workbook/add-question-section.tsx` | 253 | 11 | - |
| ts | `(anonymous)` | `cocotola-web/app/lib/observability/tracing.server.ts` | 23 | 11 | - |
| ts | `parseState` | `cocotola-web/app/lib/study-session.ts` | 105 | 11 | - |
| ts | `handleUpdateTitleIntent` | `cocotola-web/app/routes/workbooks.$workbookId.tsx` | 70 | 11 | - |
| ts | `handleCreateIntent` | `cocotola-web/app/routes/workbooks.index.tsx` | 64 | 11 | - |
| ts | `(anonymous)` | `cocotola-web/app/components/study/question-types/ordering/answer-line.tsx` | 73 | 10 | - |

## Coverage

### Layers

| Area | Layer | Covered | Lines | Coverage |
| --- | --- | --- | --- | --- |
| app | `config` | 21 | 27 | 77.8% |
| audio-generator | `config` | 8 | 22 | 36.4% |
| audio-generator | `gateway` | 151 | 252 | 59.9% |
| audio-generator | `usecase` | 151 | 155 | 97.4% |
| auth | `controller` | 844 | 1100 | 76.7% |
| auth | `domain` | 1083 | 1137 | 95.3% |
| auth | `gateway` | 1008 | 1333 | 75.6% |
| auth | `service` | 229 | 229 | 100.0% |
| auth | `usecase` | 833 | 967 | 86.1% |
| init | `config` | 50 | 56 | 89.3% |
| init | `gateway` | 116 | 159 | 73.0% |
| init | `importer` | 79 | 79 | 100.0% |
| init | `initialize` | 252 | 290 | 86.9% |
| init | `seed` | 421 | 479 | 87.9% |
| lib | `controller` | 220 | 239 | 92.1% |
| lib | `domain` | 46 | 46 | 100.0% |
| lib | `gateway` | 415 | 641 | 64.7% |
| lib | `i18n` | 3 | 3 | 100.0% |
| lib | `process` | 29 | 29 | 100.0% |
| lib | `telemetry` | 13 | 16 | 81.2% |
| lib | `usecase` | 8 | 8 | 100.0% |
| question | `config` | 19 | 25 | 76.0% |
| question | `contentjson` | 45 | 45 | 100.0% |
| question | `controller` | 1146 | 1569 | 73.0% |
| question | `domain` | 979 | 994 | 98.5% |
| question | `gateway` | 930 | 1083 | 85.9% |
| question | `initialize` | 88 | 88 | 100.0% |
| question | `service` | 352 | 376 | 93.6% |
| question | `usecase` | 1147 | 1325 | 86.6% |

### Reports

| File | Modified | Skipped |
| --- | --- | --- |
| `coverage.lcov` | 2026-11-24T00:12:05+09:00 | 0 |

## Largest files

| File | Area | Lang | Lines |
| --- | --- | --- | --- |
| `cocotola-auth/service/auth/auth.go` | auth | go | 420 |
| `cocotola-question/service/study/study.go` | question | go | 419 |
| `cocotola-web/app/routes/workbooks.$workbookId.tsx` | web | ts | 404 |
| `cocotola-web/app/routes/workbooks.$workbookId.study.tsx` | web | ts | 389 |
| `cocotola-init/initialize/initializer.go` | init | go | 377 |
| `cocotola-web/app/routes/workbooks.index.tsx` | web | ts | 377 |
| `cocotola-web/app/routes/settings.tsx` | web | ts | 376 |
| `cocotola-question/domain/question/question.go` | question | go | 368 |
| `cocotola-question/gateway/question_repository.go` | question | go | 368 |
| `cocotola-web/app/components/workbook/add-question-section.tsx` | web | ts | 354 |
