# Interactive Forma demo

Forma now includes a browser-driven office simulation with three illustrated agents, queued work, animated task execution, agent inspection, an activity feed and adjustable workstation positions.

Run the frontend independently:

```sh
cd web
npm install
npm run dev
```

Open the displayed local Vite URL and select **Office**. Agents will automatically accept mock tasks. Pause, resume, change simulation speed or reset the run. Click an avatar to inspect it, click the floor to reposition the selected agent, and open **Customize workstations** to change the room layout. Workstation positions are saved in browser local storage.

This simulation does not connect to Paseo, Overwatch, PostgreSQL or a remote backend. It intentionally exercises the visual experience before live provider integration. Agent execution history is ephemeral; only workstation coordinates persist. For the opt-in Go demo endpoints see [demo.md](demo.md).
