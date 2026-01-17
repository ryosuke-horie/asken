import ImageUpload from '@/components/client/ImageUpload'
import './page.css'

export default function Home() {
  return (
    <div className="home">
      <div className="intro">
        <h2>食事の画像から栄養素を分析</h2>
        <p>画像をアップロードするだけで、AIが自動的に食材を認識し、カロリーと栄養素を計算します。</p>
      </div>

      <ImageUpload />
    </div>
  )
}
