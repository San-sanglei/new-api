#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
File Helper for Claude Code Operations
Provides read/write/search utilities for project modification
"""
import os
import sys
import shutil
from pathlib import Path
from datetime import datetime
PROJECT_ROOT = None
def set_project_root(path):
    """Set the project root directory"""
    global PROJECT_ROOT
    PROJECT_ROOT = Path(path).resolve()
    print(f"Project root set to: {PROJECT_ROOT}")
    return True
def read_file(file_path, start_line=None, end_line=None):
    """Read file content with optional line range"""
    if PROJECT_ROOT and not Path(file_path).is_absolute():
        file_path = PROJECT_ROOT / file_path

    path = Path(file_path)
    if not path.exists():
        print(f"File not found: {path}")
        return None

    try:
        with open(path, 'r', encoding='utf-8') as f:
            content = f.read()

        if start_line is not None:
            lines = content.split('\n')
            start = max(0, start_line - 1)
            end = end_line if end_line else len(lines)
            content = '\n'.join(lines[start:end])

        print(f"Read: {path} ({len(content)} chars)")
        return content
    except Exception as e:
        print(f"Error reading: {e}")
        return None
def write_file(file_path, content, backup=True):
    """Write file content with automatic backup"""
    if PROJECT_ROOT and not Path(file_path).is_absolute():
        file_path = PROJECT_ROOT / file_path

    path = Path(file_path)

    if backup and path.exists():
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_path = str(path) + f'.backup_{timestamp}'
        shutil.copy2(path, backup_path)
        print(f"Backup created: {backup_path}")

    path.parent.mkdir(parents=True, exist_ok=True)

    try:
        with open(path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Written: {path} ({len(content)} chars)")
        return True
    except Exception as e:
        print(f"Error writing: {e}")
        return False
def find_files(pattern, directory=None, max_results=20):
    """Find files matching pattern"""
    if directory is None:
        directory = PROJECT_ROOT
    else:
        directory = Path(directory)

    matches = list(directory.rglob(pattern))
    print(f"Found {len(matches)} matches for '{pattern}':")
    for m in matches[:max_results]:
        rel_path = m.relative_to(PROJECT_ROOT) if PROJECT_ROOT else m
        print(f"  - {rel_path}")
    return matches
def grep_files(keyword, file_pattern='*.go', directory=None, max_results=20):
    """Search for keyword in files"""
    if directory is None:
        directory = PROJECT_ROOT
    else:
        directory = Path(directory)

    matches = []
    for file_path in directory.rglob(file_pattern):
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
                if keyword in content:
                    lines = content.split('\n')
                    for i, line in enumerate(lines, 1):
                        if keyword in line:
                            rel_path = file_path.relative_to(PROJECT_ROOT) if PROJECT_ROOT else file_path
                            matches.append({
                                'file': str(rel_path),
                                'line': i,
                                'content': line.strip()
                            })
                            if len(matches) >= max_results:
                                break
                    if len(matches) >= max_results:
                        break
        except:
            continue

    print(f"Found {len(matches)} matches for '{keyword}':")
    for m in matches[:15]:
        print(f"  {m['file']}:{m['line']} | {m['content'][:80]}")
    return matches
def show_tree(directory=None, max_depth=3, current_depth=0):
    """Display directory tree"""
    if directory is None:
        directory = PROJECT_ROOT
    else:
        directory = Path(directory)

    if current_depth > max_depth:
        return

    try:
        items = sorted(directory.iterdir(), key=lambda x: (not x.is_dir(), x.name))
    except PermissionError:
        return

    skip = {'.git', 'node_modules', 'dist', 'build', '.idea', '.vscode', '__pycache__'}
    items = [item for item in items if item.name not in skip]

    for i, item in enumerate(items):
        is_last = (i == len(items) - 1)
        indent = "    " * current_depth
        connector = "└── " if is_last else "├── "

        if item.is_dir():
            print(f"{indent}{connector}{item.name}/")
            if current_depth < max_depth:
                show_tree(item, max_depth, current_depth + 1)
        else:
            print(f"{indent}{connector}{item.name}")
def main():
    """Command-line interface"""
    if len(sys.argv) < 2:
        print("""File Helper for Claude Code Operations
Usage:
  python file_helper.py set_root <path>          Set project root
  python file_helper.py read <file> [start,end]  Read file (optional line range)
  python file_helper.py write <file> <content>   Write file (auto-backup)
  python file_helper.py find <pattern>           Find files
  python file_helper.py grep <keyword>           Search in files
  python file_helper.py tree [dir] [depth]       Show directory tree
Examples:
  python file_helper.py set_root D:\Documents\new-api-main
  python file_helper.py read main.go
  python file_helper.py read relay/relay.go 1,50
  python file_helper.py grep "CalculateQuota"
  python file_helper.py tree relay/adapter 3
""")
        return

    command = sys.argv[1]

    if command == "set_root":
        if len(sys.argv) < 3:
            print("Please provide project path")
            return
        set_project_root(sys.argv[2])

    elif command == "read":
        if len(sys.argv) < 3:
            print("Please provide file path")
            return
        file_path = sys.argv[2]
        lines = None
        if len(sys.argv) > 3:
            parts = sys.argv[3].split(',')
            lines = (int(parts[0]), int(parts[1]) if len(parts) > 1 else None)
        content = read_file(file_path, lines[0] if lines else None, lines[1] if lines else None)
        if content:
            print("\n--- FILE CONTENT ---\n")
            print(content)

    elif command == "write":
        if len(sys.argv) < 4:
            print("Please provide file path and content")
            return
        file_path = sys.argv[2]
        content = sys.argv[3] if len(sys.argv) > 3 else input("Enter content: ")
        write_file(file_path, content)

    elif command == "find":
        if len(sys.argv) < 3:
            print("Please provide search pattern")
            return
        find_files(sys.argv[2])

    elif command == "grep":
        if len(sys.argv) < 3:
            print("Please provide keyword")
            return
        keyword = sys.argv[2]
        pattern = sys.argv[3] if len(sys.argv) > 3 else '*.go'
        grep_files(keyword, pattern)

    elif command == "tree":
        directory = sys.argv[2] if len(sys.argv) > 2 else None
        depth = int(sys.argv[3]) if len(sys.argv) > 3 else 3
        show_tree(directory, depth)

    else:
        print(f"Unknown command: {command}")
if __name__ == "__main__":
    main()
