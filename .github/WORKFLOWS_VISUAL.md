# GitHub Actions Workflows - Visual Guide

Visual representation of all CI/CD workflows and their interactions.

## Workflow Trigger Map

```
┌─────────────────────────────────────────────────────────────────────┐
│                         GitHub Repository                            │
└─────────────────────────────────────────────────────────────────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
              ┌─────▼─────┐ ┌────▼────┐ ┌─────▼──────┐
              │   Push    │ │   PR    │ │    Tag     │
              │ to Branch │ │  Event  │ │  v*.*.*    │
              └─────┬─────┘ └────┬────┘ └─────┬──────┘
                    │            │            │
      ┌─────────────┼────────────┼────────────┼─────────────┐
      │             │            │            │             │
┌─────▼─────┐ ┌────▼────┐ ┌────▼────┐ ┌─────▼──────┐ ┌───▼────┐
│    CI     │ │ Docker  │ │   PR    │ │  Release   │ │ Deploy │
│  Pipeline │ │  Build  │ │Validate │ │  Workflow  │ │ (Manual)│
└───────────┘ └─────────┘ └─────────┘ └────────────┘ └────────┘
      │             │            │            │             │
      │             │            │            │             │
┌─────▼─────────────▼────────────▼────────────▼─────────────▼──────┐
│                     Security Scanning                             │
│              (CodeQL, Trivy, Gosec, Dependabot)                  │
└──────────────────────────────────────────────────────────────────┘
```

## CI Pipeline Flow

```
┌──────────────┐
│  Push/PR to  │
│    Branch    │
└──────┬───────┘
       │
       ▼
┌─────────────────┐
│ Changes         │
│ Detection       │◄───── Path Filters
└────────┬────────┘       (.go, .yaml, etc.)
         │
    ┌────┴────┬────────┬──────────┐
    │         │        │          │
    ▼         ▼        ▼          ▼
┌────────┐┌──────┐┌───────┐┌──────────┐
│DB Migr ││DB Load││Order  ││ Security │
│Build   ││Build  ││Food   ││  Scan    │
└───┬────┘└───┬──┘└───┬───┘└────┬─────┘
    │         │        │         │
    ▼         ▼        ▼         ▼
┌────────────────────────────────────┐
│        Test & Validate             │
│  ├─ go vet                        │
│  ├─ go fmt                        │
│  ├─ go test -race                 │
│  ├─ golangci-lint                 │
│  └─ coverage report                │
└────────────┬───────────────────────┘
             │
             ▼
      ┌─────────────┐
      │   Success   │
      │  ✅ All Pass │
      └─────────────┘
```

## Docker Build Flow

```
┌──────────────┐
│  Push to     │
│  main/tag    │
└──────┬───────┘
       │
       ▼
┌─────────────────────────────────────┐
│   Set up Multi-Platform Builder     │
│   (BuildX with QEMU)                │
└────────┬────────────────────────────┘
         │
    ┌────┴────┬────────┐
    │         │        │
    ▼         ▼        ▼
┌────────┐┌──────┐┌──────┐
│DB Migr ││DB Load││Order │
│Image   ││Image  ││Food  │
└───┬────┘└───┬──┘└───┬──┘
    │         │        │
    │    Build & Tag   │
    │  ┌───────────────┤
    │  │ latest        │
    │  │ v1.0.0       │
    │  │ v1.0         │
    │  │ v1           │
    │  │ main-sha     │
    │  └───────────────┤
    │         │        │
    ▼         ▼        ▼
┌─────────────────────────┐
│  Security Scan          │
│  (Trivy)               │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Push to Registry       │
│  ghcr.io/<user>/<mod>   │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Test Image            │
│  (order-food only)     │
└────────┬────────────────┘
         │
         ▼
      Success ✅
```

## Helm Validation Flow

```
┌──────────────┐
│  Changes to  │
│  helm/**     │
└──────┬───────┘
       │
       ▼
┌─────────────────────────┐
│  Lint All Charts        │
│  ├─ helm lint           │
│  └─ chart-testing lint  │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Template Charts        │
│  helm template          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Validate Manifests     │
│  kubeval               │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Create Kind Cluster    │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Build & Load Images    │
│  into Kind             │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Install Charts         │
│  ├─ database-migration  │
│  ├─ database-load       │
│  └─ order-food          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Verify Deployments     │
│  ├─ Check pods          │
│  ├─ Check services      │
│  └─ Test endpoints      │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Package Charts         │
│  (on main branch only)  │
└────────┬────────────────┘
         │
         ▼
      Success ✅
```

## Release Workflow

```
┌──────────────┐
│  Create Tag  │
│  v1.0.0      │
└──────┬───────┘
       │
       ▼
┌─────────────────────────┐
│  Create GitHub Release  │
│  ├─ Generate changelog  │
│  └─ Create release      │
└────────┬────────────────┘
         │
    ┌────┴────┬────────┬──────────┐
    │         │        │          │
    ▼         ▼        ▼          ▼
┌────────┐┌──────┐┌───────┐┌──────────┐
│Docker  ││Binary││ Helm  ││   Test   │
│Images  ││Builds││Charts ││  Images  │
└───┬────┘└───┬──┘└───┬───┘└────┬─────┘
    │         │        │         │
    │    Multi-Arch    │         │
    │   ┌─────────┐    │         │
    │   │ Linux   │    │         │
    │   │ macOS   │    │         │
    │   │ Windows │    │         │
    │   └─────────┘    │         │
    │         │        │         │
    ▼         ▼        ▼         ▼
┌─────────────────────────────────┐
│     Upload Release Assets       │
│  ├─ Docker images (all tags)   │
│  ├─ Binaries (all platforms)   │
│  ├─ Helm charts (.tgz)         │
│  └─ index.yaml                  │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Publish Release        │
│  ✅ Release v1.0.0      │
└─────────────────────────┘
```

## Pull Request Validation

```
┌──────────────┐
│   Create PR  │
└──────┬───────┘
       │
  ┌────┴────┬──────────┬───────────┬──────────┐
  │         │          │           │          │
  ▼         ▼          ▼           ▼          ▼
┌────┐  ┌──────┐  ┌───────┐  ┌────────┐ ┌──────┐
│Info│  │Title │  │Conflict│  │ Code   │ │ Size │
│    │  │Valid │  │ Check  │  │ Review │ │Label │
└──┬─┘  └───┬──┘  └───┬───┘  └───┬────┘ └───┬──┘
   │        │          │          │          │
   └────────┴──────────┴──────────┴──────────┘
                       │
                       ▼
            ┌─────────────────┐
            │  Post Summary   │
            │  Comment on PR  │
            └─────────────────┘
                       │
                       ▼
              ┌────────────────┐
              │   CI Pipeline  │
              │   Runs...      │
              └────────────────┘
```

## Deployment Workflow

```
┌──────────────────────┐
│  Manual Trigger      │
│  Select:             │
│  ├─ Environment      │
│  └─ Version          │
└──────┬───────────────┘
       │
       ▼
┌─────────────────────────┐
│  Configure kubectl      │
│  (environment-specific) │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Deploy via Helm        │
│  ├─ database-migration  │
│  ├─ database-load       │
│  └─ order-food          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Verify Deployment      │
│  ├─ Check pods ready    │
│  ├─ Check services      │
│  └─ Wait for rollout    │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Run Smoke Tests        │
│  ├─ Health check        │
│  ├─ API test            │
│  └─ Basic functionality │
└────────┬────────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌────────┐┌──────────┐
│Success ││  Failed  │
│✅      ││  ❌       │
└────────┘└────┬─────┘
               │
               ▼
        ┌─────────────┐
        │  Rollback   │
        │  Previous   │
        │  Version    │
        └─────────────┘
```

## Security Scanning

```
┌──────────────────────────────────┐
│  Security Scanning (Continuous)  │
└────────┬─────────────────────────┘
         │
    ┌────┴────┬────────┬───────────┐
    │         │        │           │
    ▼         ▼        ▼           ▼
┌────────┐┌──────┐┌───────┐┌────────────┐
│CodeQL  ││Gosec ││Trivy  ││Dependabot  │
│(Code)  ││(Go)  ││(Images)││(Deps)      │
└───┬────┘└───┬──┘└───┬───┘└─────┬──────┘
    │         │        │          │
    └─────────┴────────┴──────────┘
              │
              ▼
    ┌──────────────────┐
    │  SARIF Reports   │
    │  Upload to       │
    │  GitHub Security │
    └────────┬─────────┘
             │
             ▼
    ┌──────────────────┐
    │  Security Alerts │
    │  in Repository   │
    └──────────────────┘
```

## Dependabot Flow

```
┌──────────────┐
│   Weekly     │
│   Schedule   │
│   (Monday)   │
└──────┬───────┘
       │
       ▼
┌─────────────────────────┐
│  Check for Updates      │
│  ├─ Go modules          │
│  ├─ GitHub Actions      │
│  └─ Docker images       │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Create PRs             │
│  (max 5 per ecosystem)  │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  PRs Labeled            │
│  ├─ dependencies        │
│  └─ module name         │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  CI Pipeline Runs       │
│  on each PR             │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  Auto-merge             │
│  (if configured)        │
│  or await review        │
└─────────────────────────┘
```

## Workflow Dependencies

```
                    ┌──────────┐
                    │   Push   │
                    └────┬─────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    ┌────▼────┐    ┌────▼────┐    ┌────▼─────┐
    │   CI    │    │ Docker  │    │ Security │
    │Pipeline │    │  Build  │    │   Scan   │
    └────┬────┘    └────┬────┘    └────┬─────┘
         │              │              │
         └──────────────┼──────────────┘
                        │
                  All Must Pass
                        │
                        ▼
                 ┌──────────────┐
                 │  Merge to    │
                 │     Main     │
                 └──────┬───────┘
                        │
            ┌───────────┴───────────┐
            │                       │
       ┌────▼────┐            ┌────▼─────┐
       │  Helm   │            │  Deploy  │
       │Validate │            │ (Manual) │
       └─────────┘            └──────────┘
```

## Success Criteria Matrix

| Workflow | Must Pass | Can Continue on Failure |
|----------|-----------|-------------------------|
| CI Pipeline | ✅ Yes | ❌ No |
| Docker Build (PR) | ✅ Yes | ❌ No |
| Docker Build (main) | ⚠️ Warn | ✅ Yes |
| Helm Validation | ✅ Yes | ❌ No |
| PR Validation | ⚠️ Warn | ✅ Yes |
| Security Scan | ⚠️ Warn | ✅ Yes |
| CodeQL | ⚠️ Warn | ✅ Yes |

## Notification Flow

```
┌─────────────────┐
│ Workflow Event  │
└────────┬────────┘
         │
    ┌────┴────┬────────┬───────┐
    │         │        │       │
    ▼         ▼        ▼       ▼
┌──────┐ ┌───────┐ ┌─────┐ ┌──────┐
│GitHub││ Email │ │Slack││Custom│
│Checks││       │ │     ││ Hook │
└──────┘ └───────┘ └─────┘ └──────┘
    │         │        │       │
    └─────────┴────────┴───────┘
              │
              ▼
      ┌───────────────┐
      │ Team Notified │
      └───────────────┘
```

## Legend

```
┌─────────┐
│  Node   │  Process or Action
└─────────┘

    │
    ▼       Flow Direction

┌────┴────┐
│Decision │  Branch Point
└─────────┘

    ✅       Success/Pass
    ❌       Failure/Error
    ⚠️       Warning
```
