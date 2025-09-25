# Function to extract changelog content from a file
# Usage: get_changelog_content <file_path>
get_changelog_content() {
    CHANGELOG_BODY=$(cat $1)
    # Skip comments from changelog
    CHANGELOG_BODY=$(echo "$CHANGELOG_BODY" | grep -v '^<!--' | grep -v '^-->')
    # If changelog is empty, return error
    if [ -z "$CHANGELOG_BODY" ]; then
        echo "❌ Changelog is empty"
        exit 1
    fi
    echo "$CHANGELOG_BODY"
}
