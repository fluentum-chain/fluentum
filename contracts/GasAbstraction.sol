// Deploy verifier and relayer
address verifier = deployFluentumVerifier();
address relayer = deployRelayer();

// Deploy gas abstraction
GasAbstraction gasAbstraction = new GasAbstraction(FLUX_TOKEN, relayer);

// Deploy gas reimbursement
GasReimbursement gasReimbursement = new GasReimbursement(verifier, relayer); 