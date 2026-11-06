# Forma 2.5D characters

Forma renders the office, furniture, player, and agents in one PixiJS pixel-art scene. Characters use native 32×32 frames, nearest-neighbor texture sampling, integer sprite scaling, smooth grid interpolation, and Y-depth sorting. The user uses the `schmove` player sprite and agents use its satchel sprite. Side-facing art is mirrored for westward movement without creating resampled textures.

The v1 character manifest describes four idle frames and four action frames. Player action frames drive walking; satchels retain their idle loop and receive a small code-driven hover while moving or working. Seated players reuse the idle loop because the source sheet has no seated row.

Workstations expose a blocked desk footprint, chair cell, approach cell, seated direction, and reservation ID. Working agents reserve a station, follow the grid path to its approach, transition into the chair, and render a provider-state indicator above their head. A user entering an unreserved chair cell uses the same seated animation and stands on the approach cell when movement resumes.

The production sheets are copied unchanged from Jonathan Ng's `j6nca/schmove` project with the owner's authorization. Each is 128×64 with eight native 32×32 frames. Forma does not upscale or rewrite the source PNG files; Pixi applies exact integer scaling at runtime.

Run `cd web && npm install && npm run dev`. Pixi owns the complete render loop and interaction surface.
