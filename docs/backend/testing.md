# 🧪 Backend Testing Guide – CtrlB Control Plane

This guide outlines how to run unit tests and perform static checks for the backend of the CtrlB Control Plane.

---

## ✅ Run All Backend Tests

```bash
cd backend
go test ./...
```

---

## 🔍 Run Tests in a Specific Package

```bash
go test ./internal/service
```

---

## 🔎 Run a Specific Test Function

```bash
go test ./internal/service -v -run TestFunctionName
```

---

## 🧺 Run Tests With Race Detection and Coverage

```bash
go test -race -cover ./...
```

---

## 💼 Static Analysis and Linting

If you have `golangci-lint` installed:

```bash
golangci-lint run
```

Otherwise, basic checks:

```bash
go vet ./...
```
