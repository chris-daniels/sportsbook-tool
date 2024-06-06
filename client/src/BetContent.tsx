import "./BetContent.css";
import { Button, Card, Flex } from 'antd';
import React, { useEffect, useState } from 'react';


const marketMap =  new Map<string, string>([
    ["batter_hits", "Batter Hits"],
    ["pitcher_strikeouts", "Pitcher Strikeouts"],
    ["pitcher_outs", "Pitcher Outs"],
    ["batter_singles", "Batter Singles"],
]);

type OfferResult = {
    results: Offer[];
};

type Offer = {
    EventId: string;
    SportKey: string;
    EventHomeTeam: string;
    EventAwayTeam: string;
    CommenceTime: string;
    MarketKey: string;
    OutcomeName: string;
    OutcomeDesc: string;
    OutcomePoint: number;
    Price: number;
    OutlierScore: number;
};

const BetContent: React.FC = () => {
    const [data, setData] = useState<Offer[]>([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch('http://localhost:3333/offers');

                const result = await response.json() as OfferResult;
                setData(result.results);
            } catch (error) {
                console.error('Error fetching data:', error);
            }
        };

        fetchData();
    }, []);

    const getCards = () => {
        return data.map((result: Offer) => {
            // Make sure resultCommenceTime is from today
            const today = new Date();
            const commenceTime = new Date(result.CommenceTime);
            if (today.getDate() !== commenceTime.getDate()) {
                return null;
            }
            const hours = commenceTime.getHours();
            const minutes = commenceTime.getMinutes();
            const formattedTime = `${hours > 12 ? hours - 12 : hours}:${minutes < 10 ? '0' : ''}${minutes}${hours >= 12 ? 'pm' : 'am'}`;

            const marketValue = marketMap.get(result.MarketKey) || result.MarketKey;
            return (
                <Card key={result.EventId} title={`${formattedTime}: ${result.EventAwayTeam} @ ${result.EventHomeTeam}`}>
                     <h2>{marketValue}</h2>
                     <h3>{result.OutcomeDesc} - {result.OutcomeName} {result.OutcomePoint}</h3>
                     <p>Price: {result.Price}</p>
                    <p>Outlier Score: {Math.round(result.OutlierScore * 100) / 100}</p>   
                    <br />
                    <Flex justify="center">
                        <Button type="primary">Choose Bet</Button>  
                    </Flex>
                </Card>
            );
        });
    }

    return (
        <div>
            {getCards()}
        </div>
    );
};

export default BetContent;