import { HistoryDetail } from '@/types/nutrition';
import styles from './HistoryDetailView.module.css';

interface HistoryDetailViewProps {
  detail: HistoryDetail;
}

export default function HistoryDetailView({ detail }: HistoryDetailViewProps) {
  return (
    <div className={styles.container}>
      <div className={styles.imageContainer}>
        <img
          src={`http://localhost:8080/api/images/${detail.image_path.split('/').pop()}`}
          alt="食事画像"
          className={styles.image}
          onError={(e) => {
            e.currentTarget.src = '/placeholder.png';
          }}
        />
      </div>

      <div className={styles.summary}>
        <h2>合計栄養素</h2>
        <div className={styles.totalNutrients}>
          <div className={styles.nutrient}>
            <span className={styles.label}>カロリー</span>
            <span className={styles.value}>
              {Math.round(detail.total_calories)} kcal
            </span>
          </div>
          <div className={styles.nutrient}>
            <span className={styles.label}>タンパク質</span>
            <span className={styles.value}>
              {detail.total_protein.toFixed(1)} g
            </span>
          </div>
          <div className={styles.nutrient}>
            <span className={styles.label}>脂質</span>
            <span className={styles.value}>{detail.total_fat.toFixed(1)} g</span>
          </div>
          <div className={styles.nutrient}>
            <span className={styles.label}>炭水化物</span>
            <span className={styles.value}>
              {detail.total_carbohydrates.toFixed(1)} g
            </span>
          </div>
        </div>
      </div>

      <div className={styles.foods}>
        <h2>食品詳細</h2>
        <table className={styles.table}>
          <thead>
            <tr>
              <th>食品名</th>
              <th>推定量</th>
              <th>カロリー</th>
              <th>タンパク質</th>
              <th>脂質</th>
              <th>炭水化物</th>
            </tr>
          </thead>
          <tbody>
            {detail.foods.map((food, index) => (
              <tr key={index}>
                <td>{food.name}</td>
                <td>{food.estimated_amount}</td>
                <td>{Math.round(food.calories_kcal)} kcal</td>
                <td>{food.protein_g.toFixed(1)} g</td>
                <td>{food.fat_g.toFixed(1)} g</td>
                <td>{food.carbohydrates_g.toFixed(1)} g</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
