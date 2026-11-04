# Forma 3D characters (first integration pass)

Forma now renders original procedural Three.js characters above the existing PixiJS office. This replaces the old blob artwork in the office view without changing the simulation engine or the Pixi click-to-move/selection controls.

- **Agent:** rounded-square robot, recessed dark face, emissive cyan eyes, antenna, articulated limbs.
- **Player:** Explorer, with a mustard beanie, brown hair, teal hoodie, shoes and a distinct human silhouette.
- **Animation:** reusable transform rig with idle, walk, work, sit and wave poses. Agent simulation currently drives idle/walk/work; the Explorer is currently a static-position idle preview, not yet controllable.
- **Future presets:** builder, creative, captain, researcher and cozy are reserved as player customization concepts. No additional meshes are implemented yet.

The models are **procedural Three.js geometry**, not GLB exports or skinned production assets. They are editable in code and require no binary assets. The previous GLB prototypes are not used. Follow-up work should include character movement/selection for the player, real skinned GLB exports, animation blending, mesh optimization, responsive alignment and visual QA.

Run `cd web && npm install && npm run dev`. The existing Pixi canvas remains responsible for office interactions; the transparent Three.js layer has pointer events disabled.
