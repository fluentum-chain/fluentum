import { QueryClient } from "@cosmjs/stargate";
import { Tendermint34Client } from "@cosmjs/tendermint-rpc";

export class CustomFluentumQueryClient {
  private queryClient: QueryClient;

  private constructor(queryClient: QueryClient) {
    this.queryClient = queryClient;
  }

  static async connect(endpoint: string): Promise<CustomFluentumQueryClient> {
    const tmClient = await Tendermint34Client.connect(endpoint);
    const queryClient = new QueryClient(tmClient);
    return new CustomFluentumQueryClient(queryClient);
  }

  // Example custom query method
  async getCustomData(address: string): Promise<any> {
    // TODO: Implement actual query logic using generated proto types
    return {};
  }
} 