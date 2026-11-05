# Forma 2.5D characters

Forma renders the office, furniture, player, and agents in one PixiJS scene. Characters use transparent directional sprite atlases, smooth interpolation, and Y-depth sorting to create a 2.5D presentation without a second 3D renderer.

The v1 character manifest defines four directions and idle, two-frame walk, seated, and seated-work states. Ground and head anchors keep movement, seating, speech, and status indicators independent of the artwork. Replacement character packs may provide different art while preserving this contract.

Workstations expose a blocked desk footprint, chair cell, approach cell, seated direction, and reservation ID. Working agents reserve a station, follow the grid path to its approach, transition into the chair, and render a provider-state indicator above their head. A user entering an unreserved chair cell uses the same seated animation and stands on the approach cell when movement resumes.

The initial Forma agent atlas was generated with OpenAI's built-in image generation tool using an original friendly lavender robot brief, then normalized to a 1200×1200 transparent 4×5 atlas. It contains no third-party character artwork or logos.

Run `cd web && npm install && npm run dev`. Pixi owns the complete render loop and interaction surface.
