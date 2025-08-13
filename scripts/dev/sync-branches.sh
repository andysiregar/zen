#!/bin/bash

# Branch Synchronization Script
# This script helps maintain clean branch relationships

set -e

echo "🔄 Branch Synchronization Script"
echo "================================="

# Fetch latest changes
echo "📡 Fetching latest changes from remote..."
git fetch origin

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
echo "📍 Currently on branch: $CURRENT_BRANCH"

# Get commit SHAs
MAIN_SHA=$(git rev-parse origin/main)
DEVELOP_SHA=$(git rev-parse origin/develop)

echo "🌟 Main branch: $MAIN_SHA"
echo "🚀 Develop branch: $DEVELOP_SHA"

# Function to sync develop with main
sync_develop_with_main() {
    echo ""
    echo "🔄 Syncing develop with main..."
    
    git checkout develop
    git pull origin develop
    
    if git merge-base --is-ancestor $DEVELOP_SHA $MAIN_SHA; then
        echo "⬆️ Develop is behind main - fast-forwarding..."
        git reset --hard origin/main
        git push origin develop
    elif git merge-base --is-ancestor $MAIN_SHA $DEVELOP_SHA; then
        echo "✅ Develop is already ahead of main"
    else
        echo "🔀 Branches have diverged..."
        echo "Would you like to:"
        echo "1) Rebase develop onto main (recommended)"
        echo "2) Merge main into develop"
        echo "3) Reset develop to main (destructive)"
        echo "4) Cancel"
        
        read -p "Choose option (1-4): " choice
        
        case $choice in
            1)
                echo "🔄 Rebasing develop onto main..."
                if git rebase origin/main; then
                    echo "✅ Rebase successful"
                    echo "⚠️  You'll need to force push: git push origin develop --force-with-lease"
                else
                    echo "❌ Rebase failed - you may need to resolve conflicts"
                    git rebase --abort
                fi
                ;;
            2)
                echo "🔀 Merging main into develop..."
                git merge origin/main --no-edit
                git push origin develop
                ;;
            3)
                echo "⚠️  Resetting develop to main (this will lose develop-only commits)..."
                read -p "Are you sure? (y/N): " confirm
                if [[ $confirm =~ ^[Yy]$ ]]; then
                    git reset --hard origin/main
                    git push origin develop --force-with-lease
                else
                    echo "Cancelled"
                fi
                ;;
            4)
                echo "Cancelled"
                ;;
        esac
    fi
    
    # Return to original branch
    git checkout $CURRENT_BRANCH
}

# Check current state
if [[ "$MAIN_SHA" == "$DEVELOP_SHA" ]]; then
    echo "✅ Branches are already in sync!"
elif git merge-base --is-ancestor $DEVELOP_SHA $MAIN_SHA; then
    COMMITS_BEHIND=$(git rev-list --count $DEVELOP_SHA..origin/main)
    echo "⚠️  Develop is $COMMITS_BEHIND commits behind main"
    echo ""
    echo "Missing commits:"
    git log --oneline $DEVELOP_SHA..origin/main
    echo ""
    read -p "Sync develop with main? (y/N): " sync_choice
    if [[ $sync_choice =~ ^[Yy]$ ]]; then
        sync_develop_with_main
    fi
elif git merge-base --is-ancestor $MAIN_SHA $DEVELOP_SHA; then
    COMMITS_AHEAD=$(git rev-list --count origin/main..$DEVELOP_SHA)
    echo "🎯 Develop is $COMMITS_AHEAD commits ahead of main"
    echo ""
    echo "New commits in develop:"
    git log --oneline origin/main..$DEVELOP_SHA
    echo ""
    echo "✅ This is normal - develop has new changes ready for main"
else
    echo "🔀 Branches have diverged - this needs attention!"
    sync_develop_with_main
fi

echo ""
echo "✅ Script completed!"