The frontend is a React application that provides a web interface for viewing log streams from the backend agent. It allows users to see a list of available log files and select one to view its content in real-time.

## Requisites

- **Node.js**
- **pnpm** (tested on v10.24.0)

## How to run for development

1.  **Install dependencies:**
    ```sh
    # If you have not installed packages before
    pnpm install
    ```
2.  **Start the development server:**
    ```sh
    # This runs the "start" script from package.json using react-scripts
    pnpm run start
    ```
The application will be available in development mode at `http://localhost:3000`.

## Build for production

To create an optimized static build of the application, run:
`pnpm run build`

The output will be placed in the `build` directory, ready for deployment.

## TODO

* config to enable/disable autocomplete
* add config for list server
  * add right panel to show server list 
* clean up code
* add some tests
* add support for basic auth
* complete support for jwt token
* add config for protocol (https|http)
* add config for auth mode (none|basic|jwt)
* add support for dark/light theme