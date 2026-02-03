#!/usr/bin/env python3
import argparse
import json
import os
import re
import plistlib
from datetime import datetime, timezone
from pathlib import Path

DEFAULT_SKIP_DIRS = {
    ".git",
    ".svn",
    ".hg",
    "DerivedData",
    "Pods",
    "Carthage",
    "build",
    ".build",
    "node_modules",
    "dist",
    "vendor",
}

TEXT_EXTENSIONS = {
    ".swift",
    ".m",
    ".mm",
    ".h",
    ".plist",
    ".entitlements",
    ".xcprivacy",
    ".json",
    ".yml",
    ".yaml",
    ".js",
    ".ts",
    ".tsx",
    ".kt",
    ".java",
    ".rb",
    ".py",
    ".go",
    ".rs",
}

MAX_FILE_SIZE = 1024 * 1024
MAX_HITS_PER_PATTERN = 50

IOS_USAGE_KEYS = [
    "NSCameraUsageDescription",
    "NSPhotoLibraryUsageDescription",
    "NSPhotoLibraryAddUsageDescription",
    "NSMicrophoneUsageDescription",
    "NSLocationWhenInUseUsageDescription",
    "NSLocationAlwaysAndWhenInUseUsageDescription",
    "NSLocationAlwaysUsageDescription",
    "NSUserTrackingUsageDescription",
    "NSBluetoothAlwaysUsageDescription",
    "NSBluetoothPeripheralUsageDescription",
    "NSContactsUsageDescription",
    "NSCalendarsUsageDescription",
    "NSRemindersUsageDescription",
    "NSMotionUsageDescription",
    "NSHealthShareUsageDescription",
    "NSHealthUpdateUsageDescription",
]

IOS_PATTERNS = {
    "uiwebview": [r"\\bUIWebView\\b"],
    "wkwebview": [r"\\bWKWebView\\b"],
    "storekit": [
        r"\\bStoreKit\\b",
        r"\\bSKPayment\\b",
        r"\\bSKPaymentQueue\\b",
        r"\\bSKProductsRequest\\b",
    ],
    "third_party_login": [
        r"GoogleSignIn",
        r"FBSDKLoginKit",
        r"FirebaseAuth",
        r"Auth0",
        r"LINELogin",
        r"LineSDK",
        r"TwitterKit",
        r"Okta",
        r"MSAL",
        r"LoginWithAmazon",
    ],
    "tracking_ads": [
        r"\\bAdSupport\\b",
        r"\\bATTrackingManager\\b",
        r"AppTrackingTransparency",
        r"IDFA",
        r"IdentifierForAdvertising",
        r"FBSDKAppEvents",
        r"AppsFlyer",
        r"Adjust",
        r"Amplitude",
        r"Mixpanel",
        r"FirebaseAnalytics",
    ],
    "push": [r"UNUserNotificationCenter", r"registerForRemoteNotifications"],
    "background": [r"UIBackgroundModes", r"BGTaskScheduler", r"beginBackgroundTask"],
}

BACKEND_PATTERNS = {
    "external_payment": [r"stripe", r"paypal", r"braintree", r"adyen", r"checkout\\.com"],
    "subscriptions": [r"subscription", r"billing", r"trial", r"purchase"],
    "account_deletion": [r"delete account", r"account/delete", r"deactivate", r"gdpr", r"data deletion"],
}


def should_skip(path: Path) -> bool:
    for part in path.parts:
        if part in DEFAULT_SKIP_DIRS:
            return True
    return False


def iter_text_files(root: Path):
    for path in root.rglob("*"):
        if path.is_dir():
            continue
        if should_skip(path):
            continue
        if path.suffix.lower() not in TEXT_EXTENSIONS:
            continue
        try:
            if path.stat().st_size > MAX_FILE_SIZE:
                continue
        except OSError:
            continue
        yield path


def scan_patterns(root: Path, patterns: dict) -> dict:
    compiled = {key: [re.compile(p, re.IGNORECASE) for p in pats] for key, pats in patterns.items()}
    results = {key: [] for key in patterns}

    if not root.exists():
        return results

    for path in iter_text_files(root):
        try:
            content = path.read_text(errors="ignore")
        except OSError:
            continue
        for index, line in enumerate(content.splitlines()):
            for key, regexes in compiled.items():
                if len(results[key]) >= MAX_HITS_PER_PATTERN:
                    continue
                for regex in regexes:
                    match = regex.search(line)
                    if match:
                        results[key].append(
                            {
                                "path": str(path),
                                "line": index + 1,
                                "match": match.group(0),
                            }
                        )
                        break
    return results


def collect_info_plists(ios_root: Path):
    info = []
    if not ios_root.exists():
        return info
    for path in ios_root.rglob("Info.plist"):
        if should_skip(path):
            continue
        entry = {"path": str(path), "usage_descriptions": {}, "all_usage_keys": []}
        try:
            with path.open("rb") as fh:
                plist = plistlib.load(fh)
            if isinstance(plist, dict):
                entry["all_usage_keys"] = sorted([k for k in plist.keys() if k.endswith("UsageDescription")])
                for key in IOS_USAGE_KEYS:
                    if key in plist:
                        value = plist.get(key)
                        if isinstance(value, str):
                            entry["usage_descriptions"][key] = value
                        else:
                            entry["usage_descriptions"][key] = "(non-string)"
        except Exception as exc:  # pylint: disable=broad-except
            entry["error"] = str(exc)
        info.append(entry)
    return info


def collect_privacy_manifests(ios_root: Path):
    manifests = []
    if not ios_root.exists():
        return manifests
    for path in ios_root.rglob("PrivacyInfo.xcprivacy"):
        if should_skip(path):
            continue
        entry = {"path": str(path), "valid_plist": False}
        try:
            with path.open("rb") as fh:
                plistlib.load(fh)
            entry["valid_plist"] = True
        except Exception as exc:  # pylint: disable=broad-except
            entry["error"] = str(exc)
        manifests.append(entry)
    return manifests


def ensure_parent(path: Path):
    path.parent.mkdir(parents=True, exist_ok=True)


def main():
    parser = argparse.ArgumentParser(description="Scan iOS/back-end repo for App Review risk signals")
    parser.add_argument("--ios-dir", default="ios", help="iOS source root (default: ios)")
    parser.add_argument("--backend-dir", default="backend", help="Backend source root (default: backend)")
    parser.add_argument("--out", default="reports/ios-review-signals.json", help="Output JSON path")
    args = parser.parse_args()

    root = Path.cwd()
    ios_root = (root / args.ios_dir).resolve()
    backend_root = (root / args.backend_dir).resolve()
    out_path = (root / args.out).resolve()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "ios": {
            "root": str(ios_root),
            "info_plists": collect_info_plists(ios_root),
            "privacy_manifests": collect_privacy_manifests(ios_root),
            "patterns": scan_patterns(ios_root, IOS_PATTERNS),
        },
        "backend": {
            "root": str(backend_root),
            "patterns": scan_patterns(backend_root, BACKEND_PATTERNS),
        },
    }

    ensure_parent(out_path)
    out_path.write_text(json.dumps(report, indent=2, ensure_ascii=False))

    ios_manifest_count = len(report["ios"]["privacy_manifests"])
    print(f"[OK] Wrote {out_path}")
    print(f"[INFO] Privacy manifests found: {ios_manifest_count}")
    print("[INFO] Pattern hits:")
    for key, hits in report["ios"]["patterns"].items():
        print(f"  ios.{key}: {len(hits)}")
    for key, hits in report["backend"]["patterns"].items():
        print(f"  backend.{key}: {len(hits)}")


if __name__ == "__main__":
    main()
