# Cookie Authentication Guide

## Overview

Some YouTube videos require authentication (age-restricted, private videos, personal playlists). This tool supports cookie-based authentication to bypass `LOGIN_REQUIRED` errors.

## How to Get Cookies

### Method 1: Browser Extension (Recommended)

**Chrome/Edge:**
1. Install [Get cookies.txt LOCALLY](https://chrome.google.com/webstore/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc)
2. Open YouTube and log in
3. Click the extension icon
4. Save as `cookies.txt`

**Firefox:**
1. Install [cookies.txt](https://addons.mozilla.org/en-US/firefox/addon/cookies-txt/)
2. Open YouTube and log in
3. Export cookies to `cookies.txt`

### Method 2: Using yt-dlp

```bash
yt-dlp --cookies-from-browser chrome --cookies cookies.txt "https://www.youtube.com/watch?v=VIDEO_ID"
```

## Cookie File Format

The tool uses Netscape cookie file format (tab-separated):

```
# Netscape HTTP Cookie File
.youtube.com	TRUE	/	TRUE	1735689600	__Secure-1PSID	value1
.youtube.com	TRUE	/	TRUE	1735689600	__Secure-3PAPISID	value2
```

## Important Cookies

For YouTube authentication, these cookies are critical:
- `__Secure-1PSID`
- `__Secure-1PAPISID`
- `__Secure-3PSID`
- `__Secure-3PAPISID`

## Usage Examples

### Get video info with cookies
```bash
ytdl info --cookies cookies.txt "https://www.youtube.com/watch?v=VIDEO_ID"
```

### Download age-restricted video
```bash
ytdl download --cookies cookies.txt "https://www.youtube.com/watch?v=VIDEO_ID"
```

### Download private playlist
```bash
ytdl download --cookies cookies.txt "https://www.youtube.com/playlist?list=PLAYLIST_ID"
```

## Security Notes

- **Never commit your cookies file to version control**
- Cookies contain authentication tokens that give access to your account
- Add `cookies.txt` to `.gitignore`
- Cookies expire - you may need to refresh them periodically
- Store cookies securely and never share them

## Troubleshooting

### "LOGIN_REQUIRED" error persists
- Ensure you're logged into YouTube in your browser
- Re-export cookies (they may have expired)
- Verify the cookie file is in Netscape format
- Check that `__Secure-*` cookies are present

### "failed to load cookies" error
- Check the file path is correct
- Ensure the file is readable
- Verify the file format matches Netscape specification

### Videos still unavailable
- Some videos may be region-locked (cookies won't help)
- Premium-only content requires YouTube Premium cookies
- Some videos may truly be unavailable

## Example Cookie File

Here's what a valid `cookies.txt` should look like:

```
# Netscape HTTP Cookie File
# This is a generated file! Do not edit.

.youtube.com	TRUE	/	TRUE	1735689600	__Secure-1PSID	g.a000abc...
.youtube.com	TRUE	/	TRUE	1735689600	__Secure-1PAPISID	xyz123...
.youtube.com	TRUE	/	TRUE	1735689600	__Secure-3PSID	g.a000def...
.youtube.com	TRUE	/	TRUE	1735689600	__Secure-3PAPISID	abc789...
.youtube.com	TRUE	/	FALSE	0	PREF	tz=America.New_York
```

## Implementation Details

The cookie support is implemented in:
- `pkg/youtube/cookie.go` - Cookie parsing
- `pkg/youtube/watchpage.go` - Cookie jar integration
- `cmd/ytdl/info.go` - CLI flag support

All cookie handling follows TDD with comprehensive tests in `pkg/youtube/cookie_test.go`.
