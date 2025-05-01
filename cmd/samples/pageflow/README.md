This workflow demonstrates how to use Cadence workflow to power UI application through a series of state transitions.

The artificial use case started by creating a product proposal workflow (workflowID is product name), and then going through content revision, approval/rejection phases.

User interactions are implemented through workflow signals, and the latest workflow state can be fetched through workflow query.

You should be able to inspect workflow run through Cadence UI (http://localhost:8088/ if you use the standard Cadence docker image).

# Steps to run locally
Make sure you start cadence server first.

### `npm install`
Make sure you are running at least node v14 (`nvm use 14`) and install dependencies. May take a while for the first time installing these packages.

### `npm start`
This command will start the complete stack. You will need cadence server running locally separately. Alternatively you can run the commands individually below (for debugging purposes).

### `npm start-ui`
This will start the frontend website. Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

### `npm start-server`
This will start the backend node server. The frontend website will directly communicate with the node server. All traffic will then query a local instance of cadence-server.

### `npm start-worker`
This will start the pageflow worker. This is needed to be running in order for the node server to trigger signals to this worker.

## `src/server/config.js`
This config contains node server details & cadence server connection details. This can be modified if running cadence in a different port number.
