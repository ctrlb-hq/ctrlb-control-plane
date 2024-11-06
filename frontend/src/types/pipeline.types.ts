export interface Pipeline {
    id: string;
    name: string;
    type: string;
    version: string;
    hostname: string;
    platform: string;
    config: Config;
    isPipeline: boolean;
    registeredAt: string; 
  }

  // eslint-disable-next-line @typescript-eslint/no-empty-object-type
  export interface Config {
  }

  export interface ApiError {
    message: string;
  }