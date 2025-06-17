# 🛠️ Troubleshooting Guide

## 🚫 Agent Startup Issue: `Failed to register with backend server`

**Cause:** The agent cannot connect to the backend server to register.

**Quick Checks:**

- ✅ Backend is running and listening (default: `8096`)
- ✅ `BACKEND_URL` is reachable and correct (e.g. `http://localhost:8096`)
- ✅ Required env vars: `BACKEND_URL`, `PIPELINE_NAME`, `STARTED_BY`
- ✅ Agent host can reach backend (try `curl`)
- ✅ Check agent and backend logs for detailed errors

**Tip:** Use `journalctl -u <agent-service>` if running as a systemd service.

---

## 🔒 Port Binding Error: `bind: address already in use :443`

**Cause:** Port `443` is already in use by another process on the system.

**Fix:**

- ✅ Ensure no other service (like nginx or apache) is occupying port 443
- ✅ Use `sudo lsof -i :443` or `sudo netstat -tulpn | grep :443` to identify the process
