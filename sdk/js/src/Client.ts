import { StargateClient } from "@cosmjs/stargate";
import { Tendermint34Client } from "@cosmjs/tendermint-rpc";
import { CustomFluentumQueryClient } from "./custom-query";

export class FluentumClient {
  public stargateClient: StargateClient;
  public customQuery: CustomFluentumQueryClient;

  private constructor(stargateClient: StargateClient, customQuery: CustomFluentumQueryClient) {
    this.stargateClient = stargateClient;
    this.customQuery = customQuery;
  }

  static async connect(endpoint: string): Promise<FluentumClient> {
    const tmClient = await Tendermint34Client.connect(endpoint);
    const stargateClient = await StargateClient.create(tmClient);
    const customQuery = await CustomFluentumQueryClient.connect(endpoint);
    return new FluentumClient(stargateClient, customQuery);
  }
  // Add custom query methods here
} 