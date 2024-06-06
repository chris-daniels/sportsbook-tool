import "./BetContent.css";
import { Button, Card, Flex } from 'antd';
import React, { useEffect, useState } from 'react';


const marketMap =  new Map<string, string>([
    // Baseball
    ["batter_hits", "Batter Hits"],
    ["pitcher_strikeouts", "Pitcher Strikeouts"],
    ["pitcher_outs", "Pitcher Outs"],
    ["batter_singles", "Batter Singles"],
    // Baskeball
    ["player_points", "Player Points"],
    ["player_assists", "Player Assists"],
    ["player_rebounds", "Player Rebounds"],
    ["player_points_assists", "Player Points + Assists"],
    ["player_points_rebounds", "Player Points + Rebounds"],
    ["player_rebounds_assists", "Player Rebounds + Assists"],

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
                const response = await fetch('http://127.0.0.1:3333/offers');

                const result = await response.json() as OfferResult;
                setData(result.results);
            } catch (error) {
                console.error('Error fetching data:', error);
            }
        };

        fetchData();
    }, []);

    const getKey = (eventId: string, marketKey: string, outcomeDesc: string, outcomeName: string) => {
        return `${eventId}-${marketKey}-${outcomeDesc}-${outcomeName}`;
    }

    const postBet = async (key: string) => {
        // Find offer from data based on eventId
        const offer = data.find((result: Offer) => getKey(result.EventId, result.MarketKey, result.OutcomeDesc, result.OutcomeName) === key);
        if (!offer) {
            return;
        }

        // Post to /bets endpoint with offer in body
        try {
            await fetch('http://127.0.0.1:3333/bets', {
                method: 'POST',
                body: JSON.stringify(offer),
            });
        } catch (error) {
            console.error('Error posting bet:', error);
        }
    }

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
                        <Button type="primary" onClick={() => postBet(getKey(result.EventId, result.MarketKey, result.OutcomeDesc, result.OutcomeName))}>Choose Bet</Button>  
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