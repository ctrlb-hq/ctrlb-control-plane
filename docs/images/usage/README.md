# Usage screenshots

These images accompany the [usage guide](../../usage.md). The original eleven screenshots were captured on October 22, 2026, from the repository's React UI in Chromium at a 1440 × 1000 viewport. The pipeline list is cropped to its contents.

For those original screenshots, the frontend used a separate local backend and a fresh SQLite database. Pipeline creation, manifest generation, component forms, and configuration saves used the real application. Collector identities, status reports, errors, and metric samples were seeded demonstration data. No Kubernetes installation or live telemetry export was performed during that capture.

The three log-field images were extracted from the supplied `ctrlb-control-plane-demo-1.25x.mp4` video on October 22, 2026. They show the recorded demo's Attributes processor form, connected pipeline, and resulting log record in CtrlB. The frames come from 01:36.4, 01:45, and 02:14 respectively, cropped from the 1920 × 1080 recording to keep the relevant controls and fields readable. They are separate from the synthetic local fixtures; the documentation update did not rerun that cluster deployment.

The `example.com` endpoints and image tag `abc1234` are illustrative. Replace them with reachable endpoints and a published image tag or digest when following the guide. The installation token in the manifest and download commands is replaced with `<installation-token>` in the screenshot. No production credentials appear in these images.

| Image | Screen |
| --- | --- |
| `sign-in.png` | Sign in |
| `pipelines.png` | VM and Kubernetes pipeline list |
| `choose-environment.png` | Collector environment selection |
| `kubernetes-setup.png` | Kubernetes installation settings |
| `kubernetes-manifest.png` | Generated manifest and download commands |
| `pipeline-editor.png` | Shared Kubernetes pipeline in Edit Mode |
| `component-configuration.png` | Batch processor timeout set to `5s` |
| `add-log-field.png` | Attributes processor with `log_source=ctrlb_otel_collector`, from the supplied video |
| `log-field-pipeline.png` | Batch connected through Attributes to CtrlB ingestion, from the supplied video |
| `log-field-result.png` | New field in a CtrlB log record, from the supplied video |
| `review-changes.png` | Pending changes before saving |
| `collector-status.png` | Applied and failed collector examples |
| `vm-overview.png` | VM collector identity and example health history |
| `delete-pipeline.png` | Deletion dialog and uninstall commands |

## Refresh the screenshots

1. Start the frontend and an isolated backend with a fresh database, following [getting started](../../getting-started.md). Keep this database separate from an existing deployment.
2. Register a demonstration account. Create VM and Kubernetes examples using placeholder destinations. Seed synthetic collector records and samples only in the demonstration database if those screens need populated examples.
3. Follow the [usage guide](../../usage.md) through setup, editing, status, and the deletion dialog. Capture the real UI with a browser screenshot tool at the viewport above, after loading indicators and notifications have cleared. For the batch example, start with a timeout other than `5s`.
4. Remove installation tokens and other credentials from visible YAML and commands before capture. Keep manifests, browser storage, credentials, and the demonstration database out of the repository.
5. Replace the PNG files, check that text remains readable, and update the capture date and guide if the UI changed. Stop the demonstration services when finished.

To refresh the log-field images, capture the [add a field to logs example](../../usage.md#add-a-field-to-logs) and verify its result in the destination. Update the provenance above if replacing the video stills with a new recording.
