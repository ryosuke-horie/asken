import { test as base } from '@playwright/test'
import { LoginPage } from '../pages/LoginPage'
import { HomePage } from '../pages/HomePage'
import { TrainingPage } from '../pages/TrainingPage'
import { LocationsPage } from '../pages/LocationsPage'
import { EquipmentPage } from '../pages/EquipmentPage'
import { MealsPage } from '../pages/MealsPage'

type Fixtures = {
  loginPage: LoginPage
  homePage: HomePage
  trainingPage: TrainingPage
  locationsPage: LocationsPage
  equipmentPage: EquipmentPage
  mealsPage: MealsPage
}

export const test = base.extend<Fixtures>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page))
  },
  homePage: async ({ page }, use) => {
    await use(new HomePage(page))
  },
  trainingPage: async ({ page }, use) => {
    await use(new TrainingPage(page))
  },
  locationsPage: async ({ page }, use) => {
    await use(new LocationsPage(page))
  },
  equipmentPage: async ({ page }, use) => {
    await use(new EquipmentPage(page))
  },
  mealsPage: async ({ page }, use) => {
    await use(new MealsPage(page))
  },
})

export { expect } from '@playwright/test'
