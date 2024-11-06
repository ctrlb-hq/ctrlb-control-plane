export interface Agent {
    type: string;
    version: string;
    hostname: string;
    platform: string;
    configId: string;
    isPipeline: boolean;
  }

  export interface ApiError {
    message: string;
  }