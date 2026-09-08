@echo off

set SFU_PASSWORD=password

start "API" cmd /k "cd backend && go run ./cmd/api"
start "SFU" cmd /k "cd backend && go run ./cmd/sfu"
start "Frontend" cmd /k "cd frontend && npm run dev"