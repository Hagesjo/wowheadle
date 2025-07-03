# Game Concept: WoWheadle Connections Variant

## Overview

This game is inspired by the NYT "Connections" game, but uses World of Warcraft news articles and their comments as the basis for grouping. Players must group 16 comments into 4 sets of 4, where each set contains comments from the same article.

## Current Implementation

### Backend (Go)

- **HTTP Server**: Running on port 8080 with RESTful endpoints
- **Data Fetching**: Fetches articles from WoWhead RSS feed
- **Comment Processing**:
  - Removes quote tags (`[quote=...]...[/quote]` and `[quote]...[/quote]`)
  - Handles nested quotes correctly
  - Removes empty lines from comments
- **Game Logic**:
  - Progressive gameplay (submit one group at a time)
  - "One away" detection (3 out of 4 correct)
  - Tracks used comments and game state
  - Debug output showing correct solutions
  - **Color-coded difficulty system** based on quote count in comments

### Frontend (HTML/CSS/JavaScript)

- **NYT Connections Style**: Dark theme with Georgia serif font
- **Responsive Design**: Adapts to mobile and desktop screens
- **Interactive UI**:
  - 4x4 grid of comment cards
  - Click to select/deselect comments (max 4)
  - Visual feedback for correct/incorrect selections
  - Progress indicator (4 dots showing current group)
  - **Color-coded difficulty system** with NYT-style color dots
- **Game Features**:
  - Real-time status messages
  - Article links displayed in completed groups
  - "One away" feedback when close to correct grouping
  - **Proper group ordering** by difficulty (not submission order)
  - **Clickable comment expansion** in modal overlay
  - **Selection highlighting** with bright gray background and black text

### API Endpoints

- `GET /` - Serves the main game interface
- `GET /start-game` - Creates a new game with 4 articles and 16 comments
- `POST /check-solution` - Validates a group submission and returns feedback

## Game Flow

1. **Game Start**: Fetch 4 random articles, each with at least 4 comments
2. **Comment Selection**: Select 4 comments from the shuffled grid
3. **Group Submission**: Submit the group for validation
4. **Feedback**: Receive immediate feedback (correct, incorrect, or "one away")
5. **Progression**: Correct groups are moved to the appropriate group containers (ordered by difficulty)
6. **Completion**: Game ends when all 16 comments are correctly grouped

## Color-Coded Difficulty System

Groups are assigned colors based on the total number of quote tags in their comments:

- **Yellow (Easiest)**: Group with the fewest quotes
- **Green**: Second fewest quotes
- **Blue**: Second most quotes
- **Purple (Hardest)**: Group with the most quotes

Completed comment cards retain their assigned colors with matching backgrounds and black text for visibility.

## Features Implemented

- ✅ Progressive gameplay (one group at a time)
- ✅ "One away" detection and feedback (bright orange color)
- ✅ Article title and URL display in completed groups
- ✅ Quote tag removal and text cleaning
- ✅ NYT Connections-style UI with dark theme
- ✅ Responsive design for mobile and desktop
- ✅ Debug output for development
- ✅ Real-time status updates
- ✅ Visual feedback for selections and results
- ✅ **Color-coded difficulty system** with NYT-style color dots
- ✅ **Proper group ordering** by difficulty (not submission order)
- ✅ **Clickable comment expansion** in modal overlay
- ✅ **Selection highlighting** with bright gray background and black text
- ✅ **Control buttons positioned above group summaries**
- ✅ **Completed cards with matching color backgrounds and black text**

## Technical Details

- **Language**: Go (backend) + HTML/CSS/JavaScript (frontend)
- **Data Source**: WoWhead RSS feed
- **Storage**: In-memory game state (no persistence)
- **Styling**: Custom CSS matching NYT Connections aesthetic
- **Architecture**: Simple HTTP server with template rendering
- **Color Scheme**: NYT Connections colors (#ffe066 yellow, #6ee7b7 green, #60a5fa blue, #a78bfa purple)

## Future Enhancements

- Add hints or allow the player to reveal article titles
- Track player stats or streaks
- Add a timer or scoring system
- Persistent game state storage
- Multiple difficulty levels
- Social sharing features
- Sound effects for correct/incorrect submissions
- Keyboard shortcuts for selection
