@echo off
cd /d e:\硕腾网络\JT-Simulate
git add -A
git commit -m "fix: P1/P2 - event-driven register, runtime state lock, ctx propagation, rate limiter cleanup, integration tests"
git push JT-Simulate-github master
