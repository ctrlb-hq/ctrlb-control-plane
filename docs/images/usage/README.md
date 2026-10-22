# Usage screenshots

These images accompany the [usage guide](../../usage.md). The setup, batch-editing, status, and deletion screenshots were captured on October 22, 2026, from the repository's React UI in Chromium at a 1440 × 1000 viewport. The pipeline list is cropped to its contents.

For those screenshots, the frontend used a separate local backend and a fresh SQLite database. Pipeline creation, manifest generation, component forms, and configuration saves used the real application. Collector identities, status reports, errors, and metric samples were seeded demonstration data. No Kubernetes installation or live telemetry export was performed during that capture.

The three log-field images show the Attributes processor form, connected pipeline, and resulting log record in CtrlB. They are cropped from 1920 × 1080 screen captures to keep the relevant controls and fields readable. These images are separate from the synthetic local fixtures.

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
| `add-log-field.png` | Attributes processor with `log_source=ctrlb_otel_collector` |
| `log-field-pipeline.png` | Batch connected through Attributes to CtrlB ingestion |
| `log-field-result.png` | New field in a CtrlB log record |
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

To refresh the log-field images, capture the [add a field to logs example](../../usage.md#add-a-field-to-logs) and verify its result in the destination. Update the capture details above if the setup changes.
