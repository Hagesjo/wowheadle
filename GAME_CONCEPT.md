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
  - **Player-specific game state** (no shared state between sessions)
  - Debug output showing correct solutions
  - **Color-coded difficulty system** based on quote count in comments
- **API Endpoints**:
  - `GET /start-game` - Creates a new game with 4 articles and 16 comments
  - `POST /check-solution` - Validates a group submission and returns feedback
  - `GET /get-solution` - Returns the complete solution for game over scenarios

### Frontend (HTML/CSS/JavaScript)

- **NYT Connections Style**: Dark theme with Georgia serif font
- **Responsive Design**: Adapts to mobile and desktop screens
- **Interactive UI**:
  - 4x4 grid of comment cards
  - Click to select/deselect comments (max 4)
  - Visual feedback for correct/incorrect selections
  - **Mistakes remaining indicator** (4 dots showing remaining guesses)
  - **Color-coded difficulty system** with NYT-style color dots
- **Game Features**:
  - Real-time status messages
  - Article links displayed in completed groups
  - "One away" feedback when close to correct grouping
  - **Proper group ordering** by difficulty (not submission order)
  - **Clickable comment expansion** in modal overlay
  - **Selection highlighting** with bright gray background and black text
  - **Persistent game state** across browser sessions
  - **Shareable progress** in NYT Connections format
  - **Game over with complete solution reveal**

### API Endpoints

- `GET /` - Serves the main game interface
- `GET /start-game` - Creates a new game with 4 articles and 16 comments
- `POST /check-solution` - Validates a group submission and returns feedback
- `GET /get-solution` - Returns the complete solution for game over scenarios

## Game Flow

1. **Game Start**: Fetch 4 random articles, each with at least 4 comments
2. **Comment Selection**: Select 4 comments from the shuffled grid
3. **Group Submission**: Submit the group for validation
4. **Feedback**: Receive immediate feedback (correct, incorrect, or "one away")
5. **Progression**: Correct groups are moved to the appropriate group containers (ordered by difficulty)
6. **Mistakes Tracking**: Incorrect guesses reduce the mistakes remaining counter
7. **Game Over**: When 4 mistakes are made, all answers are revealed
8. **Completion**: Game ends when all 16 comments are correctly grouped

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
- ✅ **Player-specific game state** (no shared state between sessions)
- ✅ **Mistakes remaining indicator** (4 dots showing remaining guesses)
- ✅ **Persistent game state** across browser sessions using localStorage
- ✅ **Game over with complete solution reveal** when 4 mistakes are made
- ✅ **Shareable progress** in NYT Connections format with emoji squares
- ✅ **Copy to clipboard functionality** for sharing progress

## Technical Details

- **Language**: Go (backend) + HTML/CSS/JavaScript (frontend)
- **Data Source**: WoWhead RSS feed
- **Storage**:
  - Server-side: In-memory game state (shared daily games, player-specific progress)
  - Client-side: localStorage for persistent game state and guess history
- **Styling**: Custom CSS matching NYT Connections aesthetic
- **Architecture**: Simple HTTP server with template rendering
- **Color Scheme**: NYT Connections colors (#ffe066 yellow, #6ee7b7 green, #60a5fa blue, #a78bfa purple)
- **Player Isolation**: Each browser session gets a unique player ID for independent gameplay

## Future work

- Database interaction for game storages, maybe. I want it to be easy to set up on new servers with no db required for less complexity.
- Don't reuse articles from previous games
