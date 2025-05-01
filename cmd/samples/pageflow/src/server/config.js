const config = {
  server: {
    protocol: 'http',
    hostname: 'localhost',
    port: '4000',
  },
  cadence: {
    domain: 'default',
    executionStartToCloseTimeoutSeconds: 10 * 60, // workflow open for 10 minutes
    retryDelay: 100,
    retryMax: 5,
    taskList: 'pageflow',
    taskStartToCloseTimeoutSeconds: 10,
    workflowType: 'main.pageWorkflow',
  },
};

export default config;
