#!/usr/bin/env python3
"""Seeded smoke mutations in temporary copies. This is not a mutation engine."""
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent.parent
MUTATIONS = [
    (
        "development socket peer guard removed",
        "internal/auth/ingress.go",
        "if err != nil || peer == nil || !peer.IsLoopback() {",
        "if err != nil || peer == nil {",
        "./internal/auth",
        "TestDevelopmentIngressUsesSocketPeerAndBoundsRate",
    ),
    (
        "development client limit removed",
        "internal/auth/ingress.go",
        "if err := clients.Allow(r.Context(), peer.String()); err != nil {",
        "if err := clients.Allow(r.Context(), peer.String()); false && err != nil {",
        "./internal/auth",
        "TestDevelopmentIngressUsesSocketPeerAndBoundsRate",
    ),
    (
        "self-approval guard removed",
        "internal/operations/purchases.go",
        "if purchase.CreatorID == principal.ActorID {",
        "if false && purchase.CreatorID == principal.ActorID {",
        "./internal/operations",
        "TestDecisionPolicies",
    ),
    (
        "draft submission guard removed",
        "internal/operations/purchases.go",
        "if purchase.Status != domain.Draft {",
        "if false && purchase.Status != domain.Draft {",
        "./internal/operations",
        "TestSubmitPolicies",
    ),
    (
        "operation permission guard removed",
        "internal/operations/purchases.go",
        "if !principal.Can(s.policy, permission) {",
        "if false && !principal.Can(s.policy, permission) {",
        "./internal/operations",
        "TestDecisionPolicies",
    ),
    (
        "memory list tenant filter removed",
        "internal/storage/memory/store.go",
        "if purchase.TenantID == tenant {",
        "if true {",
        "./internal/operations",
        "TestOrganizationWorkflow",
    ),
    (
        "security middleware registration removed",
        "internal/app/routes.go",
        "router.Use(pluto.SecurityHeaders)",
        "// Security middleware removed by mutation.",
        "./internal/app",
        "TestRouterSecurityHeaders",
    ),
    (
        "anti-sniffing header removed",
        "internal/pluto/security.go",
        'headers.Set("X-Content-Type-Options", "nosniff")',
        "// Anti-sniffing header removed by mutation.",
        "./internal/pluto",
        "TestSecurityHeaders",
    ),
]


def execute(command, cwd, timeout):
    return subprocess.run(command, cwd=cwd, capture_output=True, text=True, timeout=timeout)


def main():
    baseline = execute(["go", "test", "./internal/operations"], ROOT, 60)
    if baseline.returncode:
        print(baseline.stdout + baseline.stderr)
        raise SystemExit("Baseline failed; mutation results would be meaningless")
    results = {"killed": 0, "survived": 0, "invalid": 0, "timed_out": 0}
    for name, filename, before, after, package, test in MUTATIONS:
        with tempfile.TemporaryDirectory(prefix="pluto-mutation-") as temporary:
            checkout = Path(temporary) / "app"
            shutil.copytree(ROOT, checkout, ignore=shutil.ignore_patterns(".git", ".env", "*.log"))
            target = checkout / filename
            source = target.read_text()
            if source.count(before) != 1:
                raise SystemExit(f"Mutation anchor must occur once: {name}")
            target.write_text(source.replace(before, after, 1))
            try:
                binary = Path(temporary) / "mutant.test"
                build = execute(["go", "test", "-c", "-o", str(binary), package], checkout, 60)
                if build.returncode:
                    outcome = "invalid"
                    print(build.stdout + build.stderr)
                else:
                    test_result = execute([str(binary), "-test.run", "^" + test + "$", "-test.timeout", "20s"], checkout, 25)
                    if "test timed out" in test_result.stdout + test_result.stderr:
                        outcome = "timed_out"
                    else:
                        outcome = "killed" if test_result.returncode else "survived"
            except subprocess.TimeoutExpired:
                outcome = "timed_out"
            results[outcome] += 1
            print(f"{outcome}: {name}")
    print("Smoke results:", results)
    if results["killed"] != len(MUTATIONS):
        raise SystemExit(1)


if __name__ == "__main__":
    main()
