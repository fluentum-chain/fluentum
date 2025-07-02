import { Tendermint34Client } from "@cosmjs/tendermint-rpc";

export class EventSubscriber {
  private tmClient: Tendermint34Client;

  private constructor(tmClient: Tendermint34Client) {
    this.tmClient = tmClient;
  }

  static async connect(endpoint: string): Promise<EventSubscriber> {
    const tmClient = await Tendermint34Client.connect(endpoint);
    return new EventSubscriber(tmClient);
  }

  subscribeNewBlocks(onBlock: (block: any) => void) {
    const subscription = this.tmClient.subscribeNewBlock();
    subscription.subscribe({
      next: onBlock,
      error: (err) => console.error("Block subscription error:", err),
      complete: () => console.log("Block subscription completed"),
    });
    return subscription;
  }

  subscribeTx(query: string, onTx: (tx: any) => void) {
    const subscription = this.tmClient.subscribeTx(query);
    subscription.subscribe({
      next: onTx,
      error: (err) => console.error("Tx subscription error:", err),
      complete: () => console.log("Tx subscription completed"),
    });
    return subscription;
  }
} 