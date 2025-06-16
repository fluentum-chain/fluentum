import fal
import numpy as np
from sklearn.ensemble import RandomForestRegressor
from sklearn.preprocessing import StandardScaler
import joblib
import json
from typing import Dict, List, Tuple, Any

class AIYieldOptimizer:
    def __init__(self):
        self.model = None
        self.scaler = StandardScaler()
        self.strategies = {}
        self.risk_profiles = {}
        
    def load_model(self, model_path: str):
        """Load the trained model and scaler"""
        self.model = joblib.load(f"{model_path}/model.joblib")
        self.scaler = joblib.load(f"{model_path}/scaler.joblib")
        
    def save_model(self, model_path: str):
        """Save the trained model and scaler"""
        joblib.dump(self.model, f"{model_path}/model.joblib")
        joblib.dump(self.scaler, f"{model_path}/scaler.joblib")
        
    def train_model(self, training_data: Dict[str, Any]):
        """Train the model on historical data"""
        X = np.array(training_data['features'])
        y = np.array(training_data['targets'])
        
        # Scale features
        X_scaled = self.scaler.fit_transform(X)
        
        # Train model
        self.model = RandomForestRegressor(
            n_estimators=100,
            max_depth=10,
            random_state=42
        )
        self.model.fit(X_scaled, y)
        
    def extract_features(
        self,
        strategy: Dict[str, Any],
        risk_profile: Dict[str, Any]
    ) -> np.ndarray:
        """Extract features for model prediction"""
        features = [
            strategy['apy'],
            strategy['risk'],
            strategy['tvl'],
            strategy['utilization'],
            risk_profile['risk_tolerance'],
            risk_profile['time_horizon'],
            risk_profile['liquidity_needs']
        ]
        return np.array(features).reshape(1, -1)
        
    def predict_strategy(
        self,
        strategy: Dict[str, Any],
        risk_profile: Dict[str, Any]
    ) -> Tuple[float, float]:
        """Predict APY and risk for a strategy"""
        features = self.extract_features(strategy, risk_profile)
        features_scaled = self.scaler.transform(features)
        
        prediction = self.model.predict(features_scaled)[0]
        predicted_apy, predicted_risk = prediction[0], prediction[1]
        
        return predicted_apy, predicted_risk
        
    def optimize_allocation(
        self,
        strategies: List[Dict[str, Any]],
        risk_profile: Dict[str, Any],
        capital: float
    ) -> List[Dict[str, Any]]:
        """Optimize capital allocation across strategies"""
        results = []
        
        for strategy in strategies:
            predicted_apy, predicted_risk = self.predict_strategy(strategy, risk_profile)
            
            # Calculate risk-adjusted return
            risk_adjusted = predicted_apy * (1 - predicted_risk)
            
            results.append({
                'strategy': strategy['id'],
                'apy': predicted_apy,
                'risk': predicted_risk,
                'risk_adjusted': risk_adjusted
            })
            
        # Sort by risk-adjusted return
        results.sort(key=lambda x: x['risk_adjusted'], reverse=True)
        
        # Calculate optimal allocations
        total_risk_adjusted = sum(r['risk_adjusted'] for r in results)
        for result in results:
            result['allocation'] = (capital * result['risk_adjusted']) / total_risk_adjusted
            
        return results
        
    def rebalance_portfolio(
        self,
        current_allocations: List[Dict[str, Any]],
        risk_profile: Dict[str, Any],
        total_capital: float
    ) -> List[Dict[str, Any]]:
        """Rebalance portfolio based on new predictions"""
        # Get current strategies
        strategies = [self.strategies[alloc['strategy_id']] for alloc in current_allocations]
        
        # Get new optimal allocations
        new_allocations = self.optimize_allocation(strategies, risk_profile, total_capital)
        
        # Calculate rebalancing moves
        rebalance_moves = []
        for i, (current, new) in enumerate(zip(current_allocations, new_allocations)):
            if abs(current['amount'] - new['allocation']) > 0.01 * total_capital:
                rebalance_moves.append({
                    'strategy_id': current['strategy_id'],
                    'current_amount': current['amount'],
                    'new_amount': new['allocation'],
                    'move_amount': new['allocation'] - current['amount']
                })
                
        return rebalance_moves

@app.function(
    machine_type="GPU-T4",
    requirements=["scikit-learn", "numpy", "joblib"]
)
def optimize_yield(
    user_risk_profile: Dict[str, Any],
    capital: float,
    strategies: List[Dict[str, Any]]
) -> Dict[str, Any]:
    """Main function to optimize yield strategies"""
    # Initialize optimizer
    optimizer = AIYieldOptimizer()
    
    # Load model
    optimizer.load_model('models/yield_model')
    
    # Get optimal allocations
    results = optimizer.optimize_allocation(strategies, user_risk_profile, capital)
    
    # Get best strategy
    best_strategy = results[0]
    
    # Execute on blockchain
    execute_strategy(best_strategy['strategy'], best_strategy['allocation'])
    
    return {
        'best_strategy': best_strategy,
        'all_allocations': results
    }

@app.function(
    machine_type="GPU-T4",
    requirements=["scikit-learn", "numpy", "joblib"]
)
def rebalance_portfolio(
    user_risk_profile: Dict[str, Any],
    current_allocations: List[Dict[str, Any]],
    total_capital: float
) -> Dict[str, Any]:
    """Rebalance portfolio based on new predictions"""
    # Initialize optimizer
    optimizer = AIYieldOptimizer()
    
    # Load model
    optimizer.load_model('models/yield_model')
    
    # Get rebalancing moves
    rebalance_moves = optimizer.rebalance_portfolio(
        current_allocations,
        user_risk_profile,
        total_capital
    )
    
    # Execute rebalancing on blockchain
    for move in rebalance_moves:
        if move['move_amount'] > 0:
            execute_strategy(move['strategy_id'], move['move_amount'])
        else:
            withdraw_strategy(move['strategy_id'], -move['move_amount'])
            
    return {
        'rebalance_moves': rebalance_moves
    }

def execute_strategy(strategy_id: str, amount: float):
    """Execute strategy on blockchain"""
    # Implementation depends on blockchain integration
    pass

def withdraw_strategy(strategy_id: str, amount: float):
    """Withdraw from strategy on blockchain"""
    # Implementation depends on blockchain integration
    pass 