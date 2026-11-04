# World and asset packs (draft v1)

The default theme is an original illustrated, friendly 2D social office. Packs must not bundle third-party game assets without permission.

A pack has a manifest with id, semantic version, minimum engine version, assets (rooms, avatar sprites, furniture, UI), animation definitions and integrity hashes. Room geometry, walkable polygons, collision boundaries, spawn points, interactive hotspots and depth layers use engine-owned logical coordinates and stable IDs. Visual textures can change without modifying those IDs.

Packs are declarative and contain no executable JavaScript. Reject absolute paths, traversal, remote script references, unsupported MIME types, oversize images and decompression bombs. Publish an explicit schema and validation CLI before supporting third-party uploads.

M4 first pass: one illustrated room, an avatar atlas, idle/walking/working/blocked animations, desk hotspots, camera pan/zoom, pointer and keyboard navigation, and accessible non-canvas controls. Future packs may replace art and animations without changing room behavior.
