# Forma 2.5D characters

Forma renders the office, furniture, player, and agents in one PixiJS pixel-art scene. Characters use transparent directional sprite atlases, nearest-neighbor texture sampling, smooth grid interpolation, and Y-depth sorting. The selectable human avatars use the same short, wide mascot proportions as robot agents while retaining distinct faces, clothes, and silhouettes.

The v1 character manifest defines four directions and idle, two-frame walk, seated, and seated-work states. Ground and head anchors keep movement, seating, speech, and status indicators independent of the artwork. Replacement character packs may provide different art while preserving this contract.

Workstations expose a blocked desk footprint, chair cell, approach cell, seated direction, and reservation ID. Working agents reserve a station, follow the grid path to its approach, transition into the chair, and render a provider-state indicator above their head. A user entering an unreserved chair cell uses the same seated animation and stands on the approach cell when movement resumes.

The initial pixel robot and two original player variants were generated with OpenAI's built-in image generation tool, then normalized to 1200×1500 transparent 4×5 atlases. The violet tech and street-inventor outfits translate broad tactical-game archetypes into original, weapon-free mascot designs without third-party logos or exact costume replicas.

Run `cd web && npm install && npm run dev`. Pixi owns the complete render loop and interaction surface.
