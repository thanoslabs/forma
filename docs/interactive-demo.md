# Interactive Forma demo

Forma now includes a browser-driven office simulation with three illustrated agents, queued work, animated task execution, agent inspection, an activity feed and adjustable workstation positions.

Run the frontend independently:

```sh
cd web
npm install
npm run dev
```

Open the displayed local Vite URL and select **Office**. Agents will automatically accept mock tasks. Pause, resume, change simulation speed or reset the run. Click an avatar to inspect it, click the floor to reposition the selected agent, and open **Customize workstations** to change the room layout. Workstation positions are saved in browser local storage.

The standalone simulation does not connect to Paseo, Overwatch, PostgreSQL or a remote backend. It intentionally exercises the visual experience without live providers. When the Forma service is available, each workspace's workstation layout is persisted through the world configuration API; standalone demo state remains browser-local. For the opt-in Go demo endpoints see [demo.md](demo.md).

The repository's GitHub Pages site builds the same React application with the `/forma/` base path. When no Forma API is available it automatically supplies three local demonstration floors, local-only chat, the elevator, and the shared street scene. Merges to `main` that change `web/` automatically rebuild and deploy the prototype through the `Forma Pages` workflow.
