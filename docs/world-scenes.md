---
title: World scenes
tags: [world, pixi, navigation]
date: 2026-11-05
---

# World scenes

Forma uses one PixiJS renderer for characters, furniture, interaction targets, and scene art. The logical world is a 35-pixel tile grid, while rendering interpolates between committed cells for smooth 2.5D movement.

Each workspace is exposed as an office floor. The elevator changes the active floor, workspace-scoped conversation, and projected agent roster without reloading the browser. Street level is a shared exterior scene; visiting it does not discard the active workspace. The building entrance opens the same destination selector used by office elevators, and each scene remembers the player's most recent position.

Office furniture contributes blocked cells, chair cells, and approach cells. Street geometry separately blocks buildings and traffic lanes while leaving sidewalks and the marked crosswalk traversable. Ambient cars are visual scene actors and do not alter provider or workspace state.

## Visual assets

The initial assets are original, replaceable production placeholders generated with OpenAI's built-in image generation tool:

- `web/public/assets/characters/forma-agent/atlas.png` and `forma-player/atlas.png` are native 128×64 sheets from `j6nca/schmove`, split into eight 32×32 frames without resampling.
- `web/public/assets/scenes/street.png` is an 860×490 New York-inspired street, park, and office entrance background.

Gameplay geometry and interaction targets are code-defined rather than baked into the images. Pixel textures use nearest-neighbor sampling, and replacement art should retain the manifest frame contract and scene dimensions or update the corresponding manifest and geometry together.
